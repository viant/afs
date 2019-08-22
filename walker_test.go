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
