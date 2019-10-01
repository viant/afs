package file

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/url"
	"os"
	"path"
	"testing"
)

func TestCreate(t *testing.T) {

	ctx := context.Background()

	tempDir := os.TempDir()

	var useCases = []struct {
		description string
		URL         string
		isDir       bool
		override    bool
		hasError    bool
	}{

		{
			description: "create file",

			URL: path.Join(tempDir, "afs", "create", "bar1.txt"),
		},
		{
			description: "override",
			override:    true,
			URL:         path.Join(tempDir, "afs", "create", "error.txt"),
		},
		{
			description: "create directory",
			isDir:       true,
			URL:         path.Join(tempDir, "afs", "create", "subdir"),
		},
		{
			description: "override directory",
			override:    true,
			URL:         path.Join(tempDir, "afs", "create", "error"),
		},
	}

	for _, useCase := range useCases {
		_ = Delete(ctx, useCase.URL)
		if useCase.override {
			isDir := useCase.isDir
			if useCase.hasError {
				isDir = !isDir
			}
			err := Create(ctx, useCase.URL, 0644, isDir)
			assert.Nil(t, err, useCase.description)
		}
		err := Create(ctx, useCase.URL, 0644, useCase.isDir)

		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}

		assert.Nil(t, err, useCase.description)
		filePath := Path(url.Path(useCase.URL))
		stat, err := os.Stat(filePath)
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, useCase.isDir, stat.IsDir())

	}

}
