package base_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/url"
	"testing"
)

func TestNew(t *testing.T) {
	ctx := context.Background()

	baseDir := "mem://localhost/tmp"
	var useCases = []struct {
		description    string
		source         string
		dest           string
		assets         []*asset.Resource
		skip           bool
		useCopyOptions bool
		skipMessage    string
	}{
		{
			description: "multi resource copy",
			source:      url.Join(baseDir, "manager/src"),
			dest:        url.Join(baseDir, "manager/dst"),

			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1", file.DefaultDirOsMode),
				asset.NewFile("folder1/asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("folder1/asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1/s1", file.DefaultDirOsMode),
				asset.NewFile("folder1/s1/asset1.txt", []byte("test 1"), 0644),
				asset.NewDir("folder1/ss", file.DefaultDirOsMode),
				asset.NewFile("folder1/s2/asset2.txt", []byte("test 2"), 0644),
				asset.NewFile("asset3.txt", []byte("test 2"), 0644),
			},
		},
	}

	mgr := mem.Singleton()

	for _, useCase := range useCases {
		service := afs.New()

		_ = asset.Cleanup(mgr, useCase.source)
		_ = asset.Cleanup(mgr, useCase.dest)

		err := asset.Create(mgr, useCase.source, useCase.assets)
		assert.Nil(t, err, useCase.description)

		err = mgr.Create(ctx, useCase.dest, 0755, true)
		assert.Nil(t, err)

		err = service.Copy(ctx, useCase.source, useCase.dest)
		assert.Nil(t, err, useCase.description)

		actuals, err := asset.Load(mgr, useCase.dest)
		assert.Nil(t, err, useCase.description)

		for _, expect := range useCase.assets {
			actual, ok := actuals[expect.Name]
			if !assert.True(t, ok, useCase.description+": "+expect.Name+fmt.Sprintf(" - actuals: %v", actuals)) {
				continue
			}
			assert.EqualValues(t, expect.Name, actual.Name, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Mode, actual.Mode, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Dir, actual.Dir, useCase.description+" "+expect.Name)
			assert.EqualValues(t, expect.Data, actual.Data, useCase.description+" "+expect.Name)
		}

		_ = service.Close(useCase.dest)
		_ = service.Close(useCase.source)
		_ = service.CloseAll()
		_ = asset.Cleanup(mgr, useCase.source)
		_ = asset.Cleanup(mgr, useCase.dest)
	}

}
