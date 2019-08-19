package mem

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDelete(t *testing.T) {

	const testContent = "this is test"
	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		isDir       bool
		hasError    bool
	}{

		{
			description: "file deletion",
			URL:         "mem:///folder/file.txt",
		},
		{
			description: "file deletion",
			URL:         "mem:///folder",
			isDir:       true,
		},
		{
			description: "file deletion error",
			URL:         "mem:///folder/file.txt",
			hasError:    true,
		},
	}

	storager := New()
	for _, useCase := range useCases {

		if !useCase.hasError {
			err := storager.Create(ctx, useCase.URL, 0744, useCase.isDir)
			assert.Nil(t, err, useCase.description)
		}
		err := storager.Delete(ctx, useCase.URL)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}

		assert.Nil(t, err, useCase.description)
	}

}
