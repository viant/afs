package mem

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {

	const testContent = "this is test"
	ctx := context.Background()
	var useCases = []struct {
		description     string
		URL             string
		uploadOptions   []storage.Option
		downloadOptions []storage.Option

		upload        bool
		downloadError bool
		readerError   bool
	}{
		{
			description: "basic asset  test",
			URL:         "mem://localhost/folder/asset.txt",
			upload:      true,
		},
		{
			description:   "download error",
			URL:           "mem://localhost/folder/asset1.txt",
			upload:        true,
			downloadError: true,
			uploadOptions: []storage.Option{
				option.Errors{
					{
						Type:  "download",
						Error: io.EOF,
					},
				},
			},
		},
		{
			description: "reader error",
			URL:         "mem://localhost/folder/asset2.txt",
			upload:      true,
			readerError: true,
			uploadOptions: []storage.Option{
				option.Errors{
					{
						Type:  "reader",
						Error: errors.New("test"),
					},
				},
			},
		},
	}

	for _, useCase := range useCases {

		manager := New()

		if useCase.upload {
			err := manager.Upload(ctx, useCase.URL, 0644, strings.NewReader(testContent), useCase.uploadOptions...)
			if !assert.Nil(t, err, useCase.description) {
				continue
			}
		}
		objects, err := manager.List(ctx, useCase.URL)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		assert.Equal(t, 1, len(objects), useCase.description)

		reader, err := manager.Open(ctx, objects[0], useCase.downloadOptions...)
		if useCase.downloadError {
			assert.NotNil(t, err, useCase.description)
			continue
		}

		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := ioutil.ReadAll(reader)
		if useCase.readerError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.Equal(t, testContent, string(actual))
	}

}
