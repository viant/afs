package afs

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"os"
	"path"
	"testing"
)

func TestService_List(t *testing.T) {
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
			baseLocation: "file://localhost/" + path.Join(baseDir, "service_walk"),
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

		actuals := map[string]bool{}

		objects, err := service.List(ctx, useCase.baseLocation, option.NewRecursive(true))
		assert.Nil(t, err)
		for _, object := range objects {
			URL := object.URL()
			relative := string(URL[len(useCase.baseLocation):])
			actuals[relative] = true
		}

		assert.Nil(t, err, useCase.description)
		for _, asset := range useCase.assets {
			_, ok := actuals[asset.Name]
			assert.True(t, ok, fmt.Sprintf(useCase.description+"  %v, %v", asset.Name, actuals))
		}

	}
}
