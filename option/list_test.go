package option_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"testing"
)

//GetListOptions returns list options
func Test_GetListOptions(t *testing.T) {

	basic, _ := matcher.NewBasic("", "", "", nil)

	var useCases = []struct {
		description string
		options     []storage.Option
		expectMatch bool
		expectPage  bool
	}{
		{
			description: "only page",
			options: []storage.Option{
				&option.Page{},
			},
			expectPage: true,
		},
		{
			description: "only matcher",
			options: []storage.Option{
				basic,
			},
			expectMatch: true,
		},
		{
			description: "only matcher",
			options: []storage.Option{
				basic.Match,
			},
			expectMatch: true,
		},
	}

	for _, useCase := range useCases {

		match, page := option.GetListOptions(useCase.options)
		var defaultMatch interface{} = option.DefaultMatch
		if useCase.expectMatch {
			assert.True(t, fmt.Sprintf("%v", defaultMatch) != fmt.Sprintf("%v", match))
			assert.NotNil(t, match)
		}
		if useCase.expectPage {
			assert.NotNil(t, page)
		}
	}
}
