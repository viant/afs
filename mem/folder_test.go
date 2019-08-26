package mem

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestFolder_Lookup(t *testing.T) {

	const testContent = "this is test"
	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		mode        os.FileMode
		content     string
		hasError    bool
		options     []storage.Option
	}{
		{
			description: "basic file test",
			URL:         "mem://localhost/folder/subfolder/file.txt",
			content:     testContent,
			mode:        os.FileMode(0744),
		},

		{
			description: "basic folder test",
			URL:         "mem://localhost/folder/subfolder/dir",
			mode:        os.FileMode(0777),
		},

		{
			description: "lookup file error",
			URL:         "mem://localhost/folder/subfolder/error.txt",
			mode:        os.FileMode(0777),
			content:     testContent,
			hasError:    true,
		},
		{
			description: "lookup folder error",
			URL:         "mem://localhost/folder/subfolder/error",
			mode:        os.FileMode(0777),
			hasError:    true,
		},
	}
	manager := newManager()
	for _, useCase := range useCases {
		if !useCase.hasError {
			if useCase.content != "" {
				err := manager.Upload(ctx, useCase.URL, useCase.mode, strings.NewReader(useCase.content), useCase.options...)
				if !assert.Nil(t, err, useCase.description) {
					continue
				}
			} else {
				err := manager.Create(ctx, useCase.URL, useCase.mode, true, useCase.options...)
				if !assert.Nil(t, err, useCase.description) {
					continue
				}
			}
		}

		baseURL, _ := url.Base(useCase.URL, Scheme)
		root := manager.Root(ctx, baseURL)
		object, err := root.Lookup(useCase.URL, 0)

		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.URL, object.URL(), useCase.description)
		assert.EqualValues(t, useCase.mode, object.Mode(), useCase.description)
		assert.EqualValues(t, useCase.content == "", object.IsDir(), useCase.description)

		if useCase.content != "" {
			file, err := root.File(object.URL())
			if assert.Nil(t, err, useCase.description) {
				continue
			}
			actual, err := ioutil.ReadAll(file.NewReader())
			if assert.Nil(t, err, useCase.description) {
				continue
			}
			assert.EqualValues(t, useCase.content, string(actual), useCase.description)
		}

	}

}
