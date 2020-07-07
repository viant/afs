package afs

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
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
			description:  "single resource test",
			baseLocation: path.Join(baseDir, "afs_new"),
			asset:        asset.NewFile("foo.txt", []byte("abc"), 0644),
		},
	}

	var err error
	for _, useCase := range useCases {
		service := New()
		_ = asset.Create(fileManager, useCase.baseLocation, []*asset.Resource{})
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

		reader, err := service.OpenURL(ctx, dest)
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

func TestService_OpenURL(t *testing.T) {
	baseDir := os.TempDir()
	ctx := context.Background()
	fileManager := file.New()

	var useCases = []struct {
		description  string
		baseLocation string
		assets       []*asset.Resource
		modifier     option.Modifier
		expect       map[string]string
	}{
		{
			description:  "location single download",
			baseLocation: path.Join(baseDir, "afs_download"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("test run by $os.User"), 0644),
				asset.NewFile("foo2.txt", []byte("test run by $os.User"), 0644),
			},
			expect: map[string]string{
				"foo1.txt": "test run by " + os.Getenv("USER"),
				"foo2.txt": "test run by $os.User",
			},
			modifier: func(info os.FileInfo, reader io.ReadCloser) (inf os.FileInfo, closer io.ReadCloser, e error) {
				if info.Name() == "foo1.txt" {
					data, err := ioutil.ReadAll(reader)
					if err != nil {
						return info, nil, err
					}
					_ = reader.Close()
					expanded := strings.Replace(string(data), "$os.User", os.Getenv("USER"), 1)
					reader = ioutil.NopCloser(strings.NewReader(expanded))
				}
				return info, reader, nil
			},
		},
	}

	for _, useCase := range useCases {
		service := New()
		err := asset.Create(fileManager, useCase.baseLocation, useCase.assets)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		for _, resource := range useCase.assets {
			URL := path.Join(useCase.baseLocation, resource.Name)
			reader, err := service.OpenURL(ctx, URL, useCase.modifier)
			if !assert.Nil(t, err, useCase.description+" "+resource.Name) {
				continue
			}
			actual, _ := ioutil.ReadAll(reader)
			expect, ok := useCase.expect[resource.Name]
			if !assert.True(t, ok, useCase.description+" "+resource.Name) {
				continue
			}
			assert.EqualValues(t, expect, actual, useCase.description+" "+resource.Name)
		}

	}
}
