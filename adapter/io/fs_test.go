package io_test

import (
	"context"
	stdio "io"
	iofs "io/fs"
	"path"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viant/afs"
	aio "github.com/viant/afs/adapter/io"
	"github.com/viant/afs/url"
)

type openCase struct {
	name      string
	openPath  string
	wantNames []string
}

func TestFS_Open(t *testing.T) {
	// Build root as file: URI to testdata
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir, _ := path.Split(thisFile)
	rootPath := path.Join(baseDir, "testdata", "fs")
	rootURI := url.ToFileURL(rootPath)

	ctx := context.Background()
	svc := afs.New()
	fsys := aio.NewFS(ctx, svc, strings.TrimRight(rootURI, "/"))

	cases := []openCase{
		{
			name:      "open root with dot",
			openPath:  ".",
			wantNames: []string{"fs", "a", "root.txt"},
		},
		{
			name:      "open root with slash",
			openPath:  "/",
			wantNames: []string{"fs", "a", "root.txt"},
		},
		{
			name:      "open subdir",
			openPath:  "a",
			wantNames: []string{"a", "alpha.txt", "sub"},
		},
		{
			name:      "open nested file returns single entry",
			openPath:  "a/sub/deep.txt",
			wantNames: []string{"deep.txt"},
		},
		{
			name:      "open root file returns single entry",
			openPath:  "root.txt",
			wantNames: []string{"root.txt"},
		},
		{
			name:      "open with cleanup - file path",
			openPath:  "a/../a/sub/./deep.txt",
			wantNames: []string{"deep.txt"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := fsys.Open(tc.openPath)
			require.NoError(t, err)
			defer f.Close()

			rdf, ok := f.(iofs.ReadDirFile)
			require.True(t, ok, "expected directory supporting ReadDirFile")
			// Read in chunks to exercise pagination
			var got []string
			for {
				ents, err := rdf.ReadDir(1)
				if err == stdio.EOF {
					break
				}
				require.NoError(t, err)
				got = append(got, ents[0].Name())
			}
			sort.Strings(got)
			sort.Strings(tc.wantNames)
			require.Equal(t, tc.wantNames, got)

			info, err := f.Stat()
			require.NoError(t, err)
			require.True(t, info.IsDir())
		})
	}

	t.Run("open missing", func(t *testing.T) {
		_, err := fsys.Open("missing.txt")
		require.Error(t, err)
	})
}
