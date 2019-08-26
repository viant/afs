package zip_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"path"
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {

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
		objects, err := service.List(ctx, useCase.URL)
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, len(useCase.expect), len(objects))
		for _, obj := range objects {
			assert.True(t, useCase.expect[obj.Name()], useCase.description+" "+obj.URL())
		}

	}
}
