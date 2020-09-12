package file

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestManager_Copy(t *testing.T) {

	const expectedContent = "this is test"

	ctx := context.Background()
	storager := New()
	tempDir := os.TempDir()

	var useCases = []struct {
		description string
		source      string
		dest        string
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
	}

	for _, useCase := range useCases {

		_ = storager.Delete(ctx, useCase.source)
		_ = storager.Delete(ctx, useCase.dest)
		if !useCase.hasError {
			err := storager.Upload(ctx, useCase.source, 0644, strings.NewReader(expectedContent))
			assert.Nil(t, err, useCase.description)
		}

		mover, _ := storager.(storage.Mover)

		err := mover.Move(ctx, useCase.source, useCase.dest)

		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}

		if !assert.Nil(t, err) {
			continue
		}

		list, err := storager.List(ctx, useCase.dest)
		if !assert.Nil(t, err) {
			continue
		}
		assert.EqualValues(t, 1, len(list), useCase.description)
		reader, err := storager.Open(ctx, list[0])
		if !assert.Nil(t, err) {
			continue
		}
		actualContent, err := ioutil.ReadAll(reader)
		_ = reader.Close()
		if !assert.Nil(t, err) {
			continue
		}
		assert.EqualValues(t, expectedContent, string(actualContent))
		err = storager.Delete(ctx, useCase.dest)
		assert.Nil(t, err)
	}

}
