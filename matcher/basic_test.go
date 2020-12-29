package matcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"path"
	"testing"
	"time"
)

func TestRegExpr_Match(t *testing.T) {
	var useCases = []struct {
		description string
		exclusion string
		prefix      string
		suffix      string
		filter      string
		location    string
		expect      bool
	}{
		{
			description: "prefix match",
			prefix:      "foo/",
			location:    "foo/abc.txt",
			expect:      true,
		},
		{
			description: "prefix no match",
			prefix:      "zfoo/",
			location:    "foo/abc.txt",
			expect:      false,
		},
		{
			description: "suffix match",
			suffix:      "txt",
			location:    "foo/abc.txt",
			expect:      true,
		},
		{
			description: "suffix no match",
			suffix:      ".abc",
			location:    "foo/abc.txt",
			expect:      false,
		},

		{
			description: "regexpr match",
			filter:      "asset\\d+\\.txt",
			location:    "asset0001.txt",
			expect:      true,
		},
		{
			description: "regexpr no match",
			filter:      "asset\\d+\\.txt",
			location:    "asset.txt",
			expect:      false,
		},

		{
			description: "regexpr no match",
			filter:      "asset\\d+\\.txt",
			location:    "asset.txt",
			expect:      false,
		},

		{
			description: "prefix and suffix match",
			prefix:      "/aa-export/v/prod/export/aa/",
			suffix:      ".gz",
			location:    "/aa-export/v/prod/export/aa/20191005_000000000000.txt.gz",
			expect:      true,
		},
		{
			description: "exclusion  - match ",
			exclusion:      ".+/data/performance/\\d+/\\d+/\\d+.+",
			location:    "/aa-export/v/prod/data/performance/2020/12/12/aa/dd/20191005_000000000000.txt.gz",
			expect:      false,
		},
		{
			description: "exclusion  - no match ",
			exclusion:      ".+/data/performance/\\d+/\\d+/\\d+.+",
			location:    "/aa-export/v/prod/data/performance/adb/12/12/aa/dd/20191005_000000000000.txt.gz",
			expect:      true,
		},
	}

	for _, useCase := range useCases {
		matcher, err := NewBasic(useCase.prefix, useCase.suffix, useCase.filter, nil)
		matcher.Exclusion = useCase.exclusion
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		parent, name := path.Split(useCase.location)
		info := file.NewInfo(name, 0, 0644, time.Now(), false)
		actual := matcher.Match(parent, info)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
