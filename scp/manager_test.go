package scp

import (
	"bytes"
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

func TestNew(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()

	keyAuth, err := LocalhostKeyAuth("")
	if err != nil {
		t.Skipf("unable to create scp auth %v", err)
	}

	fileManager := file.New()
	if !assert.Nil(t, err) {
		return
	}

	var useCases = []struct {
		description  string
		baseLocation string
		asset        *asset.Resource
	}{
		{
			description:  "location single download",
			baseLocation: path.Join(baseDir, "scp_new_01"),
			asset:        asset.NewFile("bar.txt", []byte("xyz"), 0644),
		},
	}

	for _, useCase := range useCases {
		manager := newManager(keyAuth)
		assert.EqualValues(t, Scheme, manager.Scheme(), useCase.description)
		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)

		dest := path.Join(useCase.baseLocation, useCase.asset.Name)

		_, err = manager.List(ctx, dest, 1)
		assert.NotNil(t, err, useCase.description)

		err = manager.Upload(ctx, dest, useCase.asset.Mode, bytes.NewReader(useCase.asset.Data))
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		has, err := manager.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.True(t, has, useCase.description)

		files, err := manager.List(ctx, dest, 1)
		assert.Nil(t, err, useCase.description)
		if assert.EqualValues(t, 1, len(files), useCase.description) {
			assert.EqualValues(t, useCase.asset.Name, files[0].Name())
		}

		reader, err := manager.OpenURL(ctx, dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := ioutil.ReadAll(reader)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.asset.Data, actual, useCase.description)

		err = manager.Delete(ctx, dest)
		assert.Nil(t, err)

		has, err = manager.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.False(t, has, useCase.description)

		err = manager.Close()
		assert.Nil(t, err)
	}

}

func TestManager_Walk(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()

	keyAuth, err := LocalhostKeyAuth("")
	if err != nil {
		t.Skipf("unable to create scp auth %v", err)
	}

	fileManager := file.New()
	if !assert.Nil(t, err) {
		return
	}

	var useCases = []struct {
		description  string
		baseLocation string
		assets       []*asset.Resource
	}{
		{
			description:  "multi resource walk",
			baseLocation: path.Join(baseDir, "scp_walk_01"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewDir("s1", 0744),
				asset.NewFile("s1/bar1.txt", []byte("xyz"), 0644),
				asset.NewDir("s1/s2", 0744),
				asset.NewFile("s1/s2/bar1.txt", []byte("xyz"), 0644),
				asset.NewFile("s1/bar2.txt", []byte("xyz"), 0644),
				asset.NewFile("foo2.txt", []byte("abc"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		manager := newManager()
		assert.EqualValues(t, Scheme, manager.Scheme(), useCase.description)
		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)
		err = asset.Create(fileManager, useCase.baseLocation, useCase.assets)
		assert.Nil(t, err, useCase.description)

		actuals := map[string]string{}
		err = manager.Walk(ctx, useCase.baseLocation, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
			content := ""
			if !info.IsDir() {
				if data, err := ioutil.ReadAll(reader); err == nil {
					content = string(data)
				}

			}
			actuals[path.Join(parent, info.Name())] = content
			return true, nil
		}, keyAuth)
		assert.Nil(t, err, useCase.description)

		for _, asset := range useCase.assets {
			_, ok := actuals[asset.Name]
			assert.True(t, ok, fmt.Sprintf(useCase.description+"  %v, %v", asset.Name, actuals))
		}

	}

}
