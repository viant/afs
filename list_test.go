package afs

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
)

func TestServiceList(t *testing.T) {
	baseDir := os.TempDir()
	ctx := context.Background()

	fileManager := file.New()

	var useCases = []struct {
		description  string
		baseLocation string
		assets       []*asset.Resource
	}{
		{
			description:  "l2 walk",
			baseLocation: "file://localhost" + path.Join(baseDir, "service_walk_l2"),
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
			description:  "l3 walk",
			baseLocation: "file://localhost" + path.Join(baseDir, "service_walk_l3"),
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewDir("r/s1/s1", file.DefaultDirOsMode),
				asset.NewFile("r/s1/s1/bar1.txt", []byte("xyz"), 0644),
				asset.NewDir("r/s1/s2", file.DefaultDirOsMode),
				asset.NewFile("r/s1/s2/bar1.txt", []byte("xyz"), 0644),
				asset.NewFile("r/s1/bar2.txt", []byte("xyz"), 0644),
				asset.NewFile("r/foo2.txt", []byte("abc"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		service := New()

		_ = asset.Cleanup(fileManager, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)
		err := asset.Create(fileManager, useCase.baseLocation, useCase.assets)
		assert.Nil(t, err, useCase.description)

		removePrefixLen := len(useCase.baseLocation)

		// Non-recursive List
		{
			actuals := map[string]bool{}
			objects, err := service.List(ctx, useCase.baseLocation)
			assert.Nil(t, err)
			for _, object := range objects {
				URL := object.URL()

				alsoRemoveSlash := 0
				if len(URL) > removePrefixLen {
					alsoRemoveSlash = 1
				}

				renamed := string(URL[removePrefixLen+alsoRemoveSlash:])

				actuals[renamed] = true
			}

			assert.Nil(t, err, useCase.description)
			for _, asset := range useCase.assets {
				dir := path.Dir(asset.Name)
				if dir != "." {
					continue
				}

				_, ok := actuals[asset.Name]
				assert.True(t, ok, fmt.Sprintf(useCase.description+" missing %v, %v", asset.Name, actuals))
			}

			assetMap := map[string]bool{}
			for _, asset := range useCase.assets {
				assetMap[asset.Name] = true

				// include all subdirs as visible items
				dir := path.Dir(asset.Name)
				for dir != "." && assetMap[dir] != true {
					assetMap[dir] = true
					dir = path.Dir(dir)
				}
			}

			// non-recursive list includes root named ""
			assetMap[""] = true

			for filename := range actuals {
				_, ok := assetMap[filename]
				assert.True(t, ok, fmt.Sprintf(useCase.description+" unexpected %v, %v", filename, actuals))
			}
		}

		// Recursive List
		{
			actuals := map[string]bool{}
			objects, err := service.List(ctx, useCase.baseLocation, option.NewRecursive(true))
			assert.Nil(t, err)
			for _, object := range objects {
				URL := object.URL()
				relative := string(URL[removePrefixLen+1:])
				actuals[relative] = true
			}

			assert.Nil(t, err, useCase.description)
			for _, asset := range useCase.assets {
				_, ok := actuals[asset.Name]
				assert.True(t, ok, fmt.Sprintf(useCase.description+" missing %v, %v", asset.Name, actuals))
			}

			assetMap := map[string]bool{}
			for _, asset := range useCase.assets {
				assetMap[asset.Name] = true

				// include all subdirs as visible items
				dir := path.Dir(asset.Name)
				for dir != "." && assetMap[dir] != true {
					assetMap[dir] = true
					dir = path.Dir(dir)
				}
			}

			// recursive list does not include root named ""

			for filename := range actuals {
				_, ok := assetMap[filename]
				assert.True(t, ok, fmt.Sprintf(useCase.description+" unexpected %v, %v", filename, actuals))
			}
		}

	}
}
