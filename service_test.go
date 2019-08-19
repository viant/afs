package afs

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestNewService(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()
	fileManager := file.New()

	var useCases = []struct {
		description  string
		baseLocation string
		asset        *asset.Resource
	}{
		{
			description:  "location single download",
			baseLocation: path.Join(baseDir, "afs_new"),
			asset:        asset.NewFile("foo.txt", []byte("abc"), 0644),
		},
	}

	var err error
	for _, useCase := range useCases {
		service := New()

		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)

		dest := path.Join(useCase.baseLocation, useCase.asset.Name)

		_, err = service.List(ctx, dest, 0, 1)
		assert.NotNil(t, err, useCase.description)

		err = service.Upload(ctx, dest, useCase.asset.Mode, bytes.NewReader(useCase.asset.Data))
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		has, err := service.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.True(t, has, useCase.description)

		files, err := service.List(ctx, dest, 0, 1)
		assert.Nil(t, err, useCase.description)
		if assert.EqualValues(t, 1, len(files), useCase.description) {
			assert.EqualValues(t, useCase.asset.Name, files[0].Name())
		}

		reader, err := service.DownloadWithURL(ctx, dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := ioutil.ReadAll(reader)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.asset.Data, actual, useCase.description)

		err = service.Delete(ctx, dest)
		assert.Nil(t, err)

		has, err = service.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.False(t, has, useCase.description)
	}

}
