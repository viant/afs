package afs

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestService_Walk(t *testing.T) {
	baseDir := os.TempDir()
	ctx := context.Background()

	fileManager := file.New()

	var useCases = []struct {
		description  string
		baseLocation string
		assets       []*asset.Resource
	}{
		{
			description:  "multi resource walk",
			baseLocation: path.Join(baseDir, "service_walk"),
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
			description:  "multi resource walk",
			baseLocation: path.Join(baseDir, "service_walk"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewDir("r/s1/s1", file.DefaultDirOsMode),
				asset.NewFile("r/s1/s1/bar1.txt", []byte("xyz"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		service := New()

		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)
		err := asset.Create(fileManager, useCase.baseLocation, useCase.assets)
		assert.Nil(t, err, useCase.description)

		actuals := map[string]string{}

		err = service.Walk(ctx, useCase.baseLocation, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
			content := ""
			if !info.IsDir() {
				if data, err := ioutil.ReadAll(reader); err == nil {
					content = string(data)
				}
			}
			actuals[path.Join(parent, info.Name())] = content
			return true, nil
		})

		assert.Nil(t, err, useCase.description)
		for _, asset := range useCase.assets {
			_, ok := actuals[asset.Name]
			assert.True(t, ok, fmt.Sprintf(useCase.description+"  %v, %v", asset.Name, actuals))
		}

	}
}

func TestService_Walk_StartsAtNestedDirectory(t *testing.T) {
	ctx := context.Background()
	fileManager := file.New()
	baseLocation := t.TempDir()

	assets := []*asset.Resource{
		asset.NewFile("a/f1.txt", []byte("v1"), 0644),
		asset.NewFile("a/b/f2.txt", []byte("v2"), 0644),
		asset.NewFile("a/b/c/f3.txt", []byte("v3"), 0644),
	}

	if err := asset.Create(fileManager, baseLocation, assets); !assert.Nil(t, err) {
		return
	}

	actuals := map[string]string{}
	service := New()
	err := service.Walk(ctx, path.Join(baseLocation, "a"), func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (bool, error) {
		location := path.Join(parent, info.Name())
		if info.IsDir() {
			actuals[location] = ""
			return true, nil
		}
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return false, err
		}
		actuals[location] = string(data)
		return true, nil
	})

	if !assert.Nil(t, err) {
		return
	}
	assert.EqualValues(t, map[string]string{
		"b":          "",
		"b/c":        "",
		"b/c/f3.txt": "v3",
		"b/f2.txt":   "v2",
		"f1.txt":     "v1",
	}, actuals)
}
