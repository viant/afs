package afs

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"os"
	"path"
	"testing"
)

func TestService_Move(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()

	var useCases = []struct {
		description   string
		source        string
		dest          string
		assets        []*asset.Resource
		destOptions   []storage.Option
		sourceOptions []storage.Option
	}{

		{
			description: "single file move",
			source:      path.Join(baseDir, "service_move_00/src"),
			dest:        path.Join(baseDir, "service_move_00/dst"),

			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
			},
		},

		{
			description: "mover move",
			source:      path.Join(baseDir, "service_move_01/src"),
			dest:        path.Join(baseDir, "service_move_01/dst"),

			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
			},
		},
		{
			description: "crosss storage move: file to mem",
			source:      path.Join(baseDir, "service_move_02/src"),
			dest:        "mem://" + path.Join(baseDir, "service_move_02/dst"),

			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
			},
		},

		{
			description: "crosss storage move: mem to file",
			source:      "mem://" + path.Join(baseDir, "service_move_03/src"),
			dest:        path.Join(baseDir, "service_move_03/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
			},
		},

		{
			description: "memory move",
			source:      "mem://" + path.Join(baseDir, "service_move_04/src"),
			dest:        "mem://" + path.Join(baseDir, "service_move_04/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset10.txt", []byte("test 1"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		service := New()
		srcManager, err := Manager(useCase.source, useCase.sourceOptions...)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		err = asset.Create(srcManager, useCase.source, useCase.assets)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		destManager, err := Manager(useCase.dest, useCase.destOptions...)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		_ = asset.Cleanup(destManager, useCase.dest)

		source := useCase.source
		if len(useCase.assets) == 1 {
			source = url.Join(source, useCase.assets[0].Name)
		}

		err = service.Move(ctx, source, useCase.dest, option.NewSource(useCase.sourceOptions...), option.NewDest(useCase.destOptions...))
		assert.Nil(t, err, useCase.description)

		actuals, err := asset.Load(destManager, useCase.dest)
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

		_ = service.CloseAll()
	}

}
