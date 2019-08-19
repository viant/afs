package tar

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"os"
	"path"
	"testing"
)

func TestUploader_Uploader(t *testing.T) {

	baseDir := os.TempDir()
	fileManager := file.New()

	var useCases = []struct {
		description    string
		destURL        string
		createLocation bool
		assets         []*asset.Resource
	}{

		{
			description: "multi download - unordered",
			destURL:     path.Join(baseDir, "tar_upload_01/test.tar"),
			assets: []*asset.Resource{
				asset.NewDir("test", 0744),
				asset.NewDir("test/folder2/sub", 0744),
				asset.NewFile("test/asset1.txt", []byte("xyz"), 0644),
				asset.NewFile("test/asset2.txt", []byte("xyz"), 0644),
				asset.NewDir("test/folder1", 0744),
				asset.NewFile("test/folder1/res.txt", []byte("xyz"), 0644),
				asset.NewFile("test/folder2/res1.txt", []byte("xyz"), 0644),
			},
		},

		{
			description: "multi download - link",
			destURL:     path.Join(baseDir, "tar_upload_02/test.tar"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewFile("foo2.txt", []byte("xyz"), 0644),
				asset.NewLink("sym.txt", "foo1.txt", 0644),
			},
		},
	}

	for _, useCase := range useCases {
		ctx := context.Background()
		uploader := NewBatchUploader(fileManager)
		upload, closer, err := uploader.Uploader(ctx, useCase.destURL)
		if !assert.Nil(t, err, useCase.description) {

		}
		for _, asset := range useCase.assets {
			relative, _ := path.Split(asset.Name)
			err = upload(ctx, relative, asset.Info(), asset.Reader())
			assert.Nil(t, err, useCase.description+" "+asset.Name)
		}
		err = closer.Close()
		assert.Nil(t, err, useCase.description)
		_, err = os.Stat(useCase.destURL)
		assert.Nil(t, err, useCase.description)
		_ = asset.Cleanup(fileManager, useCase.destURL)

	}

}
