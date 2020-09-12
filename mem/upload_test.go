package mem_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/option"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestStorager_Upload(t *testing.T) {

	useCases := []struct {
		description          string
		generation           *option.Generation
		hasFirstUploadError  bool
		hasSecondUploadError bool
		uploadTwice          bool
		URL                  string
	}{
		{
			description:         "new upload error",
			URL:                 "mem://localhost/mem-storager/upload/case001.txt",
			generation:          option.NewGeneration(false, 0),
			hasFirstUploadError: true,
		},
		{
			description:          "second upload error",
			URL:                  "mem://localhost/mem-storager/upload/case002.txt",
			generation:           option.NewGeneration(true, 0),
			hasSecondUploadError: true,
		},
	}

	fs := afs.New()
	ctx := context.Background()

	for _, useCase := range useCases {

		err := fs.Upload(ctx, useCase.URL, 0644, strings.NewReader("1"), useCase.generation)
		if useCase.hasFirstUploadError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		err = fs.Upload(ctx, useCase.URL, 0644, strings.NewReader("1"), useCase.generation)

		if useCase.hasSecondUploadError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

	}

}

func TestStorager_Upload_Race(t *testing.T) {

	fs := afs.New()
	ctx := context.Background()
	URL := "mem://localhost/data/test.txt"
	count := int32(0)
	waitGroup := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := fs.Upload(ctx, URL, 0644, strings.NewReader("1"), option.NewGeneration(true, 0))
			if isPreConditionError(err) {
				atomic.AddInt32(&count, 1)
			}
		}()
	}
	waitGroup.Wait()
	assert.EqualValues(t, 9, count)
	objects, _ := fs.List(context.Background(), "mem://localhost/data/")
	assert.EqualValues(t, 2, len(objects))
}

func isPreConditionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf(" %v", http.StatusPreconditionFailed))
}
