package afs

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/scp"
	"github.com/viant/afs/storage"
	"os"
	"path"
	"testing"
)

func TestService_Copy(t *testing.T) {

	baseDir := os.TempDir()
	ctx := context.Background()

	skip := false
	skipMessage := ""
	keyAuth, err := scp.LocalhostKeyAuth("")
	if err != nil {
		skip = true
		skipMessage = err.Error()
	}

	fileManager := file.New()
	var useCases = []struct {
		description    string
		sourceLocation string
		source         string
		dest           string
		assets         []*asset.Resource
		destOptions    []storage.Option
		sourceOptions  []storage.Option
		skip           bool
		useCopyOptions bool
		skipMessage    string
	}{
		{
			description: "multi resource copy",
			source:      path.Join(baseDir, "service_copy_01/src"),
			dest:        path.Join(baseDir, "service_copy_01/dst"),

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
		{
			description:    "single resource copy",
			source:         path.Join(baseDir, "service_copy_02/src/"),
			sourceLocation: path.Join(baseDir, "service_copy_02/src/asset1.txt"),
			dest:           path.Join(baseDir, "service_copy_02/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
			},
		},
		{
			description:    "file to scp",
			skip:           skip,
			skipMessage:    skipMessage,
			useCopyOptions: true,
			destOptions: []storage.Option{
				keyAuth,
			},
			source: path.Join(baseDir, "service_copy_03/src"),
			dest:   "scp://127.0.0.1:22" + path.Join(baseDir, "service_copy_03/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1", file.DefaultDirOsMode),
				asset.NewFile("folder1/asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("folder1/asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1/s1", file.DefaultDirOsMode),
				asset.NewFile("folder1/s1/asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("folder1/ss/asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1/ss/dd", file.DefaultDirOsMode),
				asset.NewFile("asset3.txt", []byte("test 2"), 0644),
			},
		},

		{
			description: "scp to file",
			skip:        skip,
			skipMessage: skipMessage,
			sourceOptions: []storage.Option{
				keyAuth,
			},
			source: "scp://127.0.0.1:22" + path.Join(baseDir, "service_copy_04/src"),
			dest:   path.Join(baseDir, "service_copy_04/dst"),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1", file.DefaultDirOsMode),
				asset.NewFile("folder1/asset1.txt", []byte("test 1"), 0644),
				asset.NewFile("folder1/asset2.txt", []byte("test 2"), 0644),
				asset.NewDir("folder1/s1", file.DefaultDirOsMode),
				asset.NewFile("folder1/s1/asset1.txt", []byte("test 1"), 0644),
				asset.NewDir("folder1/ss", file.DefaultDirOsMode),
				asset.NewFile("folder1/ss/asset2.txt", []byte("test 2"), 0644),
				asset.NewFile("asset3.txt", []byte("test 2"), 0644),
			},
		},
	}

	for _, useCase := range useCases {

		if useCase.skip {
			t.Skipf(useCase.skipMessage)
		}

		if useCase.sourceLocation == "" {
			useCase.sourceLocation = useCase.source
		}
		service := New()

		if !useCase.useCopyOptions {
			err := service.Init(ctx, useCase.source, useCase.sourceOptions...)
			assert.Nil(t, err, useCase.description)
			err = service.Init(ctx, useCase.dest, useCase.destOptions...)
			assert.Nil(t, err, useCase.description)
		}

		_ = asset.Cleanup(fileManager, useCase.source)
		_ = asset.Cleanup(fileManager, useCase.dest)
		_ = fileManager.Create(ctx, useCase.source, 0744, true)
		err = asset.Create(fileManager, useCase.source, useCase.assets)
		assert.Nil(t, err, useCase.description)

		if useCase.useCopyOptions {
			err = service.Copy(ctx, useCase.sourceLocation, useCase.dest, option.NewSource(useCase.sourceOptions...), option.NewDest(useCase.destOptions...))
		} else {
			err = service.Copy(ctx, useCase.sourceLocation, useCase.dest)
		}

		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		actuals, err := asset.Load(fileManager, useCase.dest)
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
		_ = asset.Cleanup(fileManager, useCase.source)
		_ = asset.Cleanup(fileManager, useCase.dest)
	}
}
