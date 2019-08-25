package zip_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/zip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestWalker_Walk(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	baseDir, _ := path.Split(filename)

	var useCases = []struct {
		description string
		location    string
		expect      []*asset.Resource
	}{
		{
			description: "zip walking",
			location:    path.Join(baseDir, "test/test.zip"),
			expect: []*asset.Resource{
				asset.NewDir("test", file.DefaultDirOsMode),
				asset.NewFile("test/asset1.txt", []byte("test is test\n"), 0644),
				asset.NewFile("test/asset2.txt", []byte("test is second test\n"), 0644),
				asset.NewDir("test/folder1", file.DefaultDirOsMode),
				asset.NewFile("test/folder1/res.txt", []byte("resource 1\n"), 0644),
				asset.NewDir("test/folder2", file.DefaultDirOsMode),
				asset.NewFile("test/folder2/res1.txt", []byte("resource 1\n"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		walker := zip.NewWalker(afs.New())

		ctx := context.Background()
		actuals := make(map[string]*asset.Resource)
		err := walker.Walk(ctx, useCase.location, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {

			resourceLocation := path.Join(parent, info.Name())
			linkName := ""
			if rawInfo, ok := info.(*file.Info); ok {
				linkName = rawInfo.Linkname
			}
			var data []byte
			if reader != nil {
				data, err = ioutil.ReadAll(reader)
				if err != nil {
					return false, err
				}
			}
			actuals[resourceLocation] = asset.New(parent, info.Mode(), info.IsDir(), linkName, data)
			return true, nil
		})
		assert.Nil(t, err, useCase.description)

		for _, asset := range useCase.expect {
			_, ok := actuals[asset.Name]
			assert.True(t, ok, useCase.description+" "+asset.Name)
		}
	}

}
