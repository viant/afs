package afs

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestNewFaker(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
		data        string
		errorType   string
		mode        os.FileMode
		options     []storage.Option
	}{

		{
			description: "simple upload/download",
			URL:         "s3://myBucket/folder/asset.txt",
			mode:        0644,
			data:        "this is test",
		},
		{
			description: "upload error",
			URL:         "s3://myBucket/folder/errUpload.txt",
			options: []storage.Option{
				option.NewUploadError(io.EOF),
			},
			errorType: option.ErrorTypeUpload,
			mode:      0644,
			data:      "this is test",
		},
		{
			description: "download error",
			URL:         "s3://myBucket/folder/errDownload.txt",
			options: []storage.Option{
				option.NewDownloadError(io.EOF),
			},
			errorType: option.ErrorTypeDownload,
			mode:      0644,
			data:      "this is test",
		},
		{
			description: "download error",
			URL:         "s3://myBucket/folder/errReader.txt",
			errorType:   option.ErrorTypeReader,
			options: []storage.Option{
				option.NewReaderError(fmt.Errorf("this it test")),
			},
			mode: 0644,
			data: "this is test",
		},
	}

	ctx := context.Background()
	for _, useCase := range useCases {
		service := NewFaker()

		err := service.Upload(ctx, useCase.URL, useCase.mode, strings.NewReader(useCase.data), useCase.options...)
		var reader io.ReadCloser
		if err == nil {
			reader, err = service.OpenURL(ctx, useCase.URL)
		}

		switch useCase.errorType {
		case option.ErrorTypeUpload:
			assert.NotNil(t, err)
			continue
		case option.ErrorTypeDownload:
			assert.NotNil(t, err)
			continue
		case option.ErrorTypeReader:
			_, err := ioutil.ReadAll(reader)
			assert.NotNil(t, err)
			continue
		default:

			assert.Nil(t, err, useCase.description)
			actual, err := ioutil.ReadAll(reader)
			assert.Nil(t, err, useCase.description)
			assert.EqualValues(t, useCase.data, string(actual), useCase.description)
		}

	}

}
