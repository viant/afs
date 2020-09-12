package scp

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
	"time"
)

func TestNewService(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()

	auth, err := LocalhostKeyAuth("")
	if err != nil {
		t.Skipf("unable to auth %v", err)
		return
	}
	provider := NewAuthProvider(auth, nil)
	clientConfig, err := provider.ClientConfig()
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
			baseLocation: path.Join(baseDir, "scp_service_01"),
			asset:        asset.NewFile("foo.txt", []byte("abc"), 0644),
		},
	}

	for _, useCase := range useCases {
		srv, err := NewStorager("127.0.0.1:22", time.Millisecond*15000, clientConfig)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)

		dest := path.Join(useCase.baseLocation, useCase.asset.Name)

		_, err = srv.List(ctx, dest, 0, 1)
		assert.NotNil(t, err, useCase.description)

		err = srv.Upload(ctx, dest, useCase.asset.Mode, bytes.NewReader(useCase.asset.Data))
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		has, err := srv.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.True(t, has, useCase.description)

		files, err := srv.List(ctx, dest, 0, 1)
		assert.Nil(t, err, useCase.description)
		if assert.EqualValues(t, 1, len(files), useCase.description) {
			assert.EqualValues(t, useCase.asset.Name, files[0].Name())
		}

		reader, err := srv.Open(ctx, dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := ioutil.ReadAll(reader)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.asset.Data, actual, useCase.description)

		err = srv.Delete(ctx, dest)
		assert.Nil(t, err)

		has, err = srv.Exists(ctx, dest)
		assert.Nil(t, err, useCase.description)
		assert.False(t, has, useCase.description)
	}

}
