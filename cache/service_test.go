package cache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/url"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestService_Cache(t *testing.T) {

	baseURL := file.Scheme + "://" + path.Join(os.TempDir(), "cfs")
	var useCases = []struct {
		description  string
		baseLocation string
		assets       []*asset.Resource
	}{
		{
			description:  "multi folder list",
			baseLocation: url.Join(baseURL, "cache"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewDir("s1", file.DefaultDirOsMode),
				asset.NewFile("s1/bar1.txt", []byte("xyz"), 0644),
				asset.NewDir("s1/s2", file.DefaultDirOsMode),
				asset.NewFile("s1/s2/bar1.txt", []byte("xyz"), 0644),
				asset.NewFile("s1/bar2.txt", []byte("xyz"), 0644),
				asset.NewFile("foo2.txt", []byte("abc"), 0644),
			},
		},
		{
			description:  "multi folder list",
			baseLocation: url.Join(baseURL, "nocache"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewDir("s1", file.DefaultDirOsMode),
				asset.NewFile("s1/bar1.txt", []byte("xyz"), 0644),
				asset.NewDir("s1/s2", file.DefaultDirOsMode),
				asset.NewFile("s1/s2/bar1.txt", []byte("xyz"), 0644),
				asset.NewFile("s1/bar2.txt", []byte("xyz"), 0644),
				asset.NewFile("foo2.txt", []byte("abc"), 0644),
			},
		},
	}

	fileManager := file.New()
	ctx := context.Background()
	_ = fileManager.Delete(ctx, url.Join(baseURL, CacheFile))
	for _, useCase := range useCases {
		service := Singleton(url.Join(baseURL, "cache"))
		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)
		err := asset.Create(fileManager, useCase.baseLocation, useCase.assets)

		assert.Nil(t, err, useCase.description)

		objects, err := service.List(ctx, useCase.baseLocation, option.NewRecursive(true))
		assert.Nil(t, err)
		assert.Equal(t, len(useCase.assets), len(objects), useCase.description)
		for _, object := range objects {
			exists, _ := service.Exists(ctx, object.URL())
			assert.True(t, exists, useCase.description+" / "+object.URL())
		}
		assert.Nil(t, err, useCase.description)

		for _, asset := range useCase.assets {
			if asset.Dir {
				continue
			}
			URL := url.Join(useCase.baseLocation, asset.Name)
			object, err := service.Object(ctx, URL)
			if !assert.Nil(t, err, useCase.description) {
				continue
			}
			reader, err := service.Open(ctx, object)
			if !assert.Nil(t, err, useCase.description) {
				continue
			}
			data, err := ioutil.ReadAll(reader)
			if !assert.Nil(t, err, useCase.description) {
				continue
			}
			assert.Equal(t, string(data), string(asset.Data), useCase.description+" "+URL)
		}

	}

}
