package matcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"path"
	"testing"
	"time"
)

func TestFilepath(t *testing.T) {

	var useCases = []struct {
		description string
		pattern     string
		location    string
		expect      bool
	}{
		{
			description: "ext match",
			pattern:     "*.txt",
			location:    "text.txt",
			expect:      true,
		},
		{
			description: "ext not match",
			pattern:     "*.txt",
			location:    "text.csv",
			expect:      false,
		},
		{
			description: "name match",
			pattern:     "text*",
			location:    "text.txt",
			expect:      true,
		},
		{
			description: "name not match",
			pattern:     "bar*",
			location:    "text.csv",
			expect:      false,
		},
		{
			description: "path match",
			pattern:     "bar/text*",
			location:    "bar/text.txt",
			expect:      true,
		},
		{
			description: "name not match",
			pattern:     "foo/bar*",
			location:    "foo/text.csv",
			expect:      false,
		},
		{
			description: "wildcard match",
			pattern:     "bar/t*.txt",
			location:    "bar/text.txt",
			expect:      true,
		},
		{
			description: "wildcard not match",
			pattern:     "bar/t*.txt",
			location:    "bar/abc.txt",
			expect:      false,
		},
	}

	for _, useCase := range useCases {

		matcher := Filepath(useCase.pattern)
		parent, name := path.Split(useCase.location)
		info := file.NewInfo(name, 0, 0644, time.Now(), false)
		actual := matcher(parent, info)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)

	}

}
