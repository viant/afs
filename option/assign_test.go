package option

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"io"
	"reflect"
	"testing"
)

type testFilter struct {
	message string
}

type Calc func(ops ...int) int

func TestFilter(t *testing.T) {

	testOption := &testFilter{}
	var reader io.Reader
	var calc Calc

	var useCases = []struct {
		description string
		strictMode  bool
		options     []storage.Option
		target      interface{}
		hasError    bool
	}{

		{
			description: "test interface option",
			strictMode:  true,
			options: []storage.Option{
				io.Reader(new(bytes.Buffer)),
			},
			target: &reader,
		},
		{
			description: "test function option",
			strictMode:  true,
			options: []storage.Option{
				Calc(func(ops ...int) int {
					return 0
				}),
			},
			target: &calc,
		},
		{
			description: "empty list",
			options:     []storage.Option{},
		},

		{
			description: "test struct option",
			strictMode:  true,
			options: []storage.Option{
				&testFilter{message: "abc"},
			},
			target: &testOption,
		},
	}

	for _, useCase := range useCases {
		targets := make([]interface{}, 0)
		if useCase.target != nil {
			targets = append(targets, useCase.target)
		}
		targets = append(targets, &FilterMode{useCase.strictMode})
		_, err := Assign(useCase.options, targets...)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		if useCase.target != nil {
			assert.EqualValues(t, fmt.Sprintf("%v", reflect.ValueOf(useCase.target).Elem().Interface()), fmt.Sprintf("%v", useCase.options[0]), useCase.description)
		}
	}

}
