package afs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

func TestService_Uploader(t *testing.T) {

	baseDir := os.TempDir()
	var useCases = []struct {
		description string
		destURL     string
		assets      []*asset.Resource
	}{

		{
			description: "batch asset upload",
			destURL:     path.Join(baseDir, "service_uploader_01/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 1"), 0644),
			},
		},
		{
			description: "batch multi level  asset upload",
			destURL:     path.Join(baseDir, "service_uploader_02/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1", file.DefaultDirOsMode),
				asset.NewFile("folder1/asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("folder1/asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1/s1", file.DefaultDirOsMode),
				asset.NewFile("folder1/s1/asset1.txt", []byte("test 1"), 0644),
				asset.NewDir("folder1/ss", file.DefaultDirOsMode),
				asset.NewFile("folder1/ss/asset2.txt", []byte("test 2"), 0644),
				asset.NewFile("asset3.txt", []byte("test 2"), 0644),
			},
		},
	}

	ctx := context.Background()
	for _, useCase := range useCases {
		service := New()

		destManager, err := Manager(useCase.destURL)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		_ = asset.Cleanup(destManager, useCase.destURL)
		assert.Nil(t, err, useCase.description)
		upload, closer, err := service.Uploader(ctx, useCase.destURL)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		for _, asset := range useCase.assets {
			relative := ""
			var reader io.Reader
			if strings.Contains(asset.Name, "/") {
				relative, _ = path.Split(asset.Name)
			}
			if !asset.Dir {
				reader = bytes.NewReader(asset.Data)
			}
			err = upload(ctx, relative, asset.Info(), reader)
			assert.Nil(t, err, useCase.description+" "+asset.Name)
		}
		_ = closer.Close()

		actuals, err := asset.Load(destManager, useCase.destURL)
		if !assert.Nil(t, err) {
			continue
		}
		for _, expect := range useCase.assets {
			actual, ok := actuals[expect.Name]
			if !assert.True(t, ok, useCase.description+": "+expect.Name+fmt.Sprintf(" - actuals: %v", actuals)) {
				continue
			}
			assert.EqualValues(t, expect.Name, actual.Name, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Mode, actual.Mode, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Dir, actual.Dir, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Data, actual.Data, useCase.description+" "+expect.Name)
		}
		_ = asset.Cleanup(destManager, useCase.destURL)
		_ = service.Close(useCase.destURL)
	}

}
