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
		options     []storage.Option
		target      interface{}
		assigned    bool
	}{

		{
			description: "test interface option",
			options: []storage.Option{
				io.Reader(new(bytes.Buffer)),
			},
			assigned: true,
			target:   &reader,
		},
		{
			description: "test function option",
			assigned:    true,
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
			assigned:    true,
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

		_, assigned := Assign(useCase.options, targets...)
		if !assert.EqualValues(t, assigned, useCase.assigned, useCase.description) {
			continue
		}
		if !assigned {
			continue
		}
		if useCase.target != nil {
			assert.EqualValues(t, fmt.Sprintf("%v", reflect.ValueOf(useCase.target).Elem().Interface()), fmt.Sprintf("%v", useCase.options[0]), useCase.description)
		}
	}

}
