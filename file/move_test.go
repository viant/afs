package file

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestMove(t *testing.T) {
	const expectedContent = "this is test"

	ctx := context.Background()

	tempDir := os.TempDir()
	manager := New()
	var useCases = []struct {
		description string
		source      string
		dest        string
		options     []storage.Option
		hasError    bool
	}{

		{
			description: "simple copy",
			source:      path.Join(tempDir, "afs", "copy", "src", "bar1.txt"),
			dest:        path.Join(tempDir, "afs", "copy", "URL", "bar1.txt"),
		},
		{
			description: "simple copy - missing source",
			hasError:    true,
			source:      path.Join(tempDir, "afs", "copy", "src", "bar2.txt"),
			dest:        path.Join(tempDir, "afs", "copy", "URL", "bar2.txt"),
		},
		{
			description: "invalid option",
			options: []storage.Option{
				storage.Option(1),
			},
			hasError: true,
			source:   path.Join(tempDir, "afs", "copy", "src", "bar1.txt"),
			dest:     path.Join(tempDir, "afs", "copy", "URL", "bar1.txt"),
		},
	}

	for _, useCase := range useCases {

		_ = Delete(ctx, useCase.source)
		_ = Delete(ctx, useCase.dest)
		if !useCase.hasError {
			err := Upload(ctx, useCase.source, 0644, strings.NewReader(expectedContent))
			assert.Nil(t, err, useCase.description)
		}

		mover, _ := manager.(storage.Mover)
		err := mover.Move(ctx, useCase.source, useCase.dest, useCase.options...)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}

		if !assert.Nil(t, err) {
			continue
		}

		reader, err := OpenURL(ctx, useCase.dest)
		if assert.Nil(t, err) {
			continue
		}
		actualContent, err := ioutil.ReadAll(reader)
		_ = reader.Close()
		if assert.Nil(t, err) {
			continue
		}
		assert.EqualValues(t, expectedContent, string(actualContent))

		_, URLPath := url.Base(useCase.source, Scheme)
		_, err = os.Stat(URLPath)
		assert.NotNil(t, err, useCase.description)

	}
}
