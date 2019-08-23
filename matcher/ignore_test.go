package matcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"path"
	"testing"
	"time"
)

func TestIgnore_Match(t *testing.T) {

	var ignoreList = []string{
		"foa",
		"/bar",
		"pre*",
		"*suf",
		"baz*qux",
		"abc/",
		"**/cde",
		"efk/**",
		"go.mod",
		"e2e/**",
		"deploy/**",
		"manager/**",
	}

	var useCases = []struct {
		description string
		ignoreList  []string
		location    string
		expect      bool
	}{

		{
			ignoreList: []string{
				"/bar",
				"pre*",
			},
			description: "ignored by rule pre*",
			location:    "m/preaaa.text",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "no rules apply, do not ignored",
			location:    "yyy/nbn/kkk.txt",
			expect:      true,
		},
		{
			ignoreList:  ignoreList,
			description: "no rules apply, do not ignored",
			location:    "kkk.txt",
			expect:      true,
		},

		{
			ignoreList:  ignoreList,
			description: "no rules apply, do not ignored",
			location:    "aaa.b/ccc/ddd/eee/kkk.txt",
			expect:      true,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule foa",
			location:    "manager/app/foa",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule foa",
			location:    "manager/app/foa",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule foa",
			location:    "foa",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule /bar",
			location:    "bar/foo.text",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "do not ignored by rule /bar",
			location:    "bar",
			expect:      true,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule pre*",
			location:    "pre",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule *suf",
			location:    "m/test.suf",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule *suf",
			location:    "test2.suf",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule baz*qux",
			location:    "m/bazaaaqux",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule baz*qux",
			location:    "bazaaaqux",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "do not ignored by rule baz*qux",
			location:    "bazaaaqux.test",
			expect:      true,
		},

		{
			ignoreList:  ignoreList,
			description: "do not ignored by rule baz*qux",
			location:    "test.bazaaaqux",
			expect:      true,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule abc/",
			location:    "abc/aaa.txt",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule abc/",
			location:    "abc",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "do not ignored by rule **/cde",
			location:    "a/cde/aaa.txt",
			expect:      true,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule **/cde",
			location:    "a/cde",
			expect:      false,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule efk/**",
			location:    "efk/aaa.txt",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "do not ignored by rule efk/**",
			location:    "a/efk",
			expect:      true,
		},
		{
			ignoreList:  ignoreList,
			description: "ignored by rule go.mod",
			location:    "vendor/gopkg.in/yaml.v2/go.mod",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule e2e/**",
			location:    "e2e/app_cf.yaml",
			expect:      false,
		},

		{
			ignoreList:  ignoreList,
			description: "ignored by rule e2e/**",
			location:    "e2e/vvv/app_cf.yaml",
			expect:      false,
		},
	}

	for _, useCase := range useCases {
		matcher, err := NewIgnore(useCase.ignoreList)
		assert.Nil(t, err, useCase.description)
		parent, name := path.Split(useCase.location)
		info := file.NewInfo(name, 0, 0644, time.Now(), false)
		actual := matcher.Match(parent, info)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
