package tar_test

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/afs"
)

func TestRewriteRootFileDoesNotAddVirtualRootEntry(t *testing.T) {
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	require.NoError(t, writer.WriteHeader(&tar.Header{Name: "shared/", Mode: 0755, Typeflag: tar.TypeDir}))
	for _, entry := range []struct{ name, body string }{{"main.yaml", "root\n"}, {"shared/source.yaml", "before\n"}} {
		require.NoError(t, writer.WriteHeader(&tar.Header{Name: entry.name, Mode: 0644, Size: int64(len(entry.body)), Typeflag: tar.TypeReg}))
		_, err := writer.Write([]byte(entry.body))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	ctx := context.Background()
	location := "mem://localhost/tar-root-rewrite/input.tar"
	backing := afs.New()
	require.NoError(t, backing.Upload(ctx, location, 0600, bytes.NewReader(archive.Bytes())))
	t.Cleanup(func() { _ = backing.Delete(ctx, "mem://localhost/tar-root-rewrite") })
	require.NoError(t, afs.New().Upload(ctx, "mem:localhost/tar-root-rewrite/input.tar/tar://localhost/shared/source.yaml", 0644, bytes.NewBufferString("after\n")))
	updated, err := backing.DownloadWithURL(ctx, location)
	require.NoError(t, err)
	reader := tar.NewReader(bytes.NewReader(updated))
	entries := map[string]string{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.NotContains(t, entries, header.Name, "duplicate archive entry")
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		entries[header.Name] = string(body)
	}
	require.Equal(t, map[string]string{"main.yaml": "root\n", "shared/": "", "shared/source.yaml": "after\n"}, entries)
}
