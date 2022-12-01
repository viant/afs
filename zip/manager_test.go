package zip_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"path"
	"runtime"
	"testing"
)

type useCaseFn func(s afs.Service, ctx context.Context, url string) ([]storage.Object, error)

func TestNew(t *testing.T) {
	testCases(t, func(service afs.Service, ctx context.Context, url string) ([]storage.Object, error) {
			return service.List(ctx, url)
	})
}


func testCases(t *testing.T, callList useCaseFn) {
	_, filename, _, _ := runtime.Caller(0)
	baseDir, _ := path.Split(filename)
	ctx := context.Background()

	var useCases = []struct {
		description string
		URL         string
		expect      map[string]bool
	}{
		{
			description: "list war classes",
			URL:         fmt.Sprintf("file:%v/test/app.war/zip://localhost/WEB-INF/classes", baseDir),
			expect: map[string]bool{
				"classes":           true,
				"HelloWorld.class":  true,
				"config.properties": true,
			},
		},

		{
			description: "list war classes",
			URL:         fmt.Sprintf("file:%v/test/app.war/zip://localhost/WEB-INF/classes/config.properties", baseDir),
			expect: map[string]bool{
				"config.properties": true,
			},
		},
	}

	for _, useCase := range useCases {
		service := afs.New()
		objects, err := callList(service, ctx, useCase.URL)
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, len(useCase.expect), len(objects))
		for _, obj := range objects {
			assert.True(t, useCase.expect[obj.Name()], useCase.description+" "+obj.URL())
		}

	}
}

func TestNoCache(t *testing.T) {
	testCases(t, func(service afs.Service, ctx context.Context, url string) ([]storage.Object, error) {
			return service.List(ctx, url, &option.NoCache{Source: option.NoCacheBaseURL})
	})
}
