package matcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"path"
	"testing"
	"time"
)

func timePtr(time time.Time) *time.Time {
	return &time
}

func TestModification_Match(t *testing.T) {

	var useCases = []struct {
		description string
		prefix      string
		suffix      string
		filter      string
		before      *time.Time
		after       *time.Time
		modTime     *time.Time
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
			description: "before time match",
			modTime:     timePtr(time.Now()),
			before:      timePtr(time.Now().Add(time.Hour)),
			location:    "asset.txt",
			expect:      true,
		},
		{
			description: "before no time match",
			modTime:     timePtr(time.Now()),
			before:      timePtr(time.Now().Add(-time.Hour)),
			location:    "asset.txt",
			expect:      false,
		},
		{
			description: "after time match",
			modTime:     timePtr(time.Now()),
			after:       timePtr(time.Now().Add(time.Hour)),
			location:    "asset.txt",
			expect:      false,
		},
		{
			description: "after no time match",
			modTime:     timePtr(time.Now()),
			after:       timePtr(time.Now().Add(-time.Hour)),
			location:    "asset.txt",
			expect:      true,
		},
	}

	for _, useCase := range useCases {
		basic, err := NewBasic(useCase.prefix, useCase.suffix, useCase.filter, nil)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		matcher := NewModification(useCase.before, useCase.after, basic.Match)
		parent, name := path.Split(useCase.location)
		modTime := time.Now()
		if useCase.modTime != nil {
			modTime = *useCase.modTime
		}
		info := file.NewInfo(name, 0, 0644, modTime, false)
		actual := matcher.Match(parent, info)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
