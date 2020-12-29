package sync

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"testing"
)

func TestCounter_Increment(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
	}{
		{
			description: "memory sync counter",
			URL:         "mem://localhost/counter/case001/data.cnt",
		},
	}

	fs := afs.New()
	for _, useCase := range useCases {
		ctx := context.Background()
		counter := NewCounter(useCase.URL, fs)
		for i := 0; i < 10; i++ {
			count, err := counter.Increment(ctx)
			if assert.Nil(t, err, useCase.description) {
				continue
			}
			assert.EqualValues(t, i+1, count)
		}
		for i := 10; i >= 0; i-- {
			count, err := counter.Decrement(ctx)
			if assert.Nil(t, err, useCase.description) {
				continue
			}
			assert.EqualValues(t, i-1, count)
		}
	}
}
