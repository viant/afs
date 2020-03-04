package mem_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/option"
	"strings"
	"testing"
)

func TestStorager_Upload(t *testing.T) {



	useCases := []struct{
		description        string
		generation         *option.Generation
		hasFirstUploadError bool
		hasSecondUploadError bool
		uploadTwice        bool
		URL                string

	}{
		{
			description:"new upload error",
			URL:"mem://localhost/mem-storager/upload/case001.txt",
			generation:option.NewGeneration(false, 0),
			hasFirstUploadError: true,
		},
		{
			description:"second upload error",
			URL:"mem://localhost/mem-storager/upload/case002.txt",
			generation:option.NewGeneration(true, 0),
			hasSecondUploadError: true,
		},

	}


	fs := afs.New()
	ctx := context.Background()

	for _, useCase:= range useCases {

		err := fs.Upload(ctx, useCase.URL, 0644, strings.NewReader("1"),  useCase.generation)
		if useCase.hasFirstUploadError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if ! assert.Nil(t, err, useCase.description) {
			continue
		}


		err = fs.Upload(ctx, useCase.URL, 0644, strings.NewReader("1"),  useCase.generation)

		if useCase.hasSecondUploadError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if ! assert.Nil(t, err, useCase.description) {
			continue
		}



	}

}