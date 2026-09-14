package io_test

import (
	"context"
	stdio "io"
	iofs "io/fs"
	"os"
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
	wantData  *string
	wantError bool
}

func TestFS_Open(t *testing.T) {
	// Build in-memory filesystem rooted at mem://localhost/fs
	ctx := context.Background()
	svc := afs.New()
	rootURI := "mem://localhost/fs"

	// Create structure:
	// /fs
	//   /a
	//     /sub
	//       deep.txt
	//     alpha.txt
	//   root.txt
	require.NoError(t, svc.Create(ctx, url.Join(rootURI, "a"), os.FileMode(0o755), true))
	require.NoError(t, svc.Create(ctx, url.Join(rootURI, "a", "sub"), os.FileMode(0o755), true))
	require.NoError(t, svc.Upload(ctx, url.Join(rootURI, "root.txt"), os.FileMode(0o644), strings.NewReader("root file\n")))
	require.NoError(t, svc.Upload(ctx, url.Join(rootURI, "a", "alpha.txt"), os.FileMode(0o644), strings.NewReader("alpha")))
	require.NoError(t, svc.Upload(ctx, url.Join(rootURI, "a", "sub", "deep.txt"), os.FileMode(0o644), strings.NewReader("deep")))

	fsys := aio.NewFS(ctx, svc, strings.TrimRight(rootURI, "/"))

	cases := []openCase{
		{
			name:      "open root with dot",
			openPath:  ".",
			wantNames: []string{"a", "root.txt"},
		},
		{
			name:      "open root with slash",
			openPath:  "/",
			wantError: true,
			wantNames: []string{"a", "root.txt"},
		},
		{
			name:      "open subdir",
			openPath:  "a",
			wantNames: []string{"alpha.txt", "sub"},
		},
		{
			name:     "open nested file returns single entry",
			openPath: "a/sub/deep.txt",
			wantData: ptr("deep"),
		},
		{
			name:     "open root file returns single entry",
			openPath: "root.txt",
			wantData: ptr("root file\n"),
		},
		{
			name:      "open with cleanup - file path",
			openPath:  "a/../a/sub/./deep.txt",
			wantError: true,
			wantNames: []string{"deep.txt"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := fsys.Open(tc.openPath)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.wantData != nil {
				defer f.Close()
				data, err := stdio.ReadAll(f)
				require.NoError(t, err)
				require.Equal(t, *tc.wantData, string(data))
				info, err := f.Stat()
				require.NoError(t, err)
				require.False(t, info.IsDir())
				return
			}
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

func ptr(value string) *string { return &value }
