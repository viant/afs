package tar

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestNewStorager(t *testing.T) {

	mgr := mem.New()

	var useCases = []struct {
		description string
		baseURL     string
		init        []*asset.Resource
		resource    *asset.Resource
		options     []storage.Option
	}{
		{
			description: "single resource archive",
			baseURL:     "mem:localhost/my001.tar/tar://localhost/",
			resource:    asset.NewFile("folder1/res1.txt", []byte("this is test"), 0644),
		},

		//{
		//	description: "single resource archive",
		//	baseURL:     "mem:localhost/my002.tar/tar://localhost/",
		//	resource:    asset.NewFile("folder1/res1.txt", []byte("this is test"), 0644),
		//	init: []*asset.Resource{
		//		asset.NewDir("folder2", 0750),
		//		asset.NewDir("folder2/sub", 0750),
		//		asset.NewFile("folder2/sub/res2.txt", []byte("xyz"), 0644),
		//		asset.NewFile("folder2/sub/res3.txt", []byte("xyz"), 0644),
		//		asset.NewFile("folder2/res4.txt", []byte("xyz"), 0644),
		//	},
		//},
	}

	ctx := context.Background()
	for _, useCase := range useCases {
		storager, err := newStorager(ctx, useCase.baseURL, mgr)
		assert.Nil(t, err, useCase.description)
		if len(useCase.init) > 0 {
			upload, closer, _ := storager.Uploader(ctx, "")
			for _, resource := range useCase.init {
				err := upload(ctx, resource.Name, resource.Info(), resource.Reader())
				assert.Nil(t, err, useCase.description)
			}
			err = closer.Close()
			assert.Nil(t, err, useCase.description)
		}

		ok, _ := storager.Exists(ctx, useCase.resource.Name)
		assert.False(t, ok, useCase.description)

		parent, _ := path.Split(useCase.resource.Name)
		objects, err := storager.List(ctx, parent)

		assert.EqualValues(t, 0, len(objects), useCase.description)
		if len(useCase.init) == 0 && !assert.NotNil(t, err, useCase.description) {
			continue
		}

		err = storager.Upload(ctx, useCase.resource.Name, useCase.resource.Mode, bytes.NewReader(useCase.resource.Data), useCase.options...)
		assert.Nil(t, err, useCase.description)
		objects, err = storager.List(ctx, parent)
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, 2, len(objects), useCase.description)

		ok, _ = storager.Exists(ctx, useCase.resource.Name)
		assert.True(t, ok, useCase.description)

		reader, err := storager.Open(ctx, useCase.resource.Name)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		data, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		_ = reader.Close()
		assert.EqualValues(t, useCase.resource.Data, string(data), useCase.description)

		visits := 0
		err = storager.Walk(ctx, useCase.resource.Name, func(parent string, info os.FileInfo, reader io.Reader) (b bool, e error) {
			visits++
			return true, nil
		})
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, 1, visits, useCase.description)

		err = storager.Delete(ctx, useCase.resource.Name)
		assert.Nil(t, err)

	}

}
