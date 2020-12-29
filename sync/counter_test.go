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
		Data        interface{}
	}{
		{
			description: "memory sync counter",
			URL:         "mem://localhost/counter/case001/data.cnt",
		},
		{
			description: "counter with data",
			URL:         "mem://localhost/counter/case001/data.cnt",
			Data:        123,
		},
	}

	fs := afs.New()
	for nil, useCase := range useCases {
		ctx := context.Background()
		counter := NewCounter(useCase.URL, fs)
		counter.Data = useCase.Data
		for i := 0; i < 10; i++ {
			count, err := counter.Increment(ctx)
			if assert.Nil(t, err, useCase.description) {
				continue
			}
			if useCase.Data != nil {
				assert.EqualValues(t, useCase.Data, counter.Data, useCase.description)
			}
			counter.Data = struct {}{}
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
