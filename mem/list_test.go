package mem

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"strings"
	"testing"
)

func TestList(t *testing.T) {

	const testContent = "this is test"
	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		files       []string
		folders     []string
		isDir       bool
		reset       bool
		hasError    bool
	}{
		{
			description: "basic folder test",
			URL:         "mem://localhost/folder/subfolder",
			isDir:       true,
			files:       []string{"file1.txt", "file2.csv"},
			folders:     []string{"dir1", "dir2"},
		},

		{
			description: "basic file test",
			URL:         "mem://localhost/folder/file2.txt",
		},

		{
			description: "basic file test",
			hasError:    true,
			URL:         "mem://localhost/folder/file3.txt",
		},
	}

	var err error

	for _, useCase := range useCases {
		storager := New()
		baseURL, _ := url.Base(useCase.URL, Scheme)
		var expect = make(map[string]bool)
		expect[useCase.URL] = true
		if len(useCase.files) > 0 {
			for i := range useCase.files {
				URL := url.Join(baseURL, useCase.files[i])
				expect[URL] = false
				err = storager.Upload(ctx, URL, 0644, strings.NewReader(testContent))
				assert.Nil(t, err, useCase.description)
			}
		}
		if len(useCase.folders) > 0 {
			for i := range useCase.folders {
				URL := url.Join(baseURL, useCase.folders[i])
				expect[URL] = true
				err = storager.Create(ctx, URL, 0644, true)
				assert.Nil(t, err, useCase.description)
			}
		}
		if !useCase.hasError {
			err = storager.Create(ctx, useCase.URL, file.DefaultDirOsMode, true)
			assert.Nil(t, err, useCase.description)
		}

		objects, err := storager.List(ctx, useCase.URL)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		for i := range objects {
			isDir, has := expect[objects[i].URL()]
			assert.True(t, has, useCase.description+" / "+objects[i].URL())
			assert.Equal(t, isDir, objects[i].IsDir(), useCase.description+" / "+objects[i].URL())
		}
	}
}
