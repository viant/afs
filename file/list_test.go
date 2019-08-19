package file

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestManager_List(t *testing.T) {

	baseDir := os.TempDir()
	fileManager := New()

	var useCases = []struct {
		description  string
		location     string
		baseLocation string
		assets       map[string]string
	}{
		{
			description:  "location download",
			location:     path.Join(baseDir, "file_list/"),
			baseLocation: path.Join(baseDir, "file_list"),
		},
	}

	ctx := context.Background()
	for _, useCase := range useCases {
		manager := New()
		assert.EqualValues(t, Scheme, manager.Scheme(), useCase.description)
		_ = fileManager.Delete(ctx, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)

		for name, content := range useCase.assets {
			filename := path.Join(baseDir, name)
			_ = ioutil.WriteFile(filename, []byte(content), 0744)
		}

		actuals := map[string]string{}
		objects, err := manager.List(ctx, useCase.location)
		assert.Nil(t, err, useCase.description)
		for _, object := range objects {
			content := ""
			if !object.IsDir() {
				reader, err := manager.Download(ctx, object)
				assert.Nil(t, err)
				if data, err := ioutil.ReadAll(reader); err == nil {
					content = string(data)
				}
			}
			actuals[object.Name()] = content
		}

		for name := range useCase.assets {
			_, ok := actuals[name]
			assert.True(t, ok, fmt.Sprintf(useCase.description+"  %v, %v", name, actuals))
		}

	}

}
