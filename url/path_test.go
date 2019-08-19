package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPath(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
		expect      string
	}{
		{
			description: "raw path",
			URL:         "/tmp/foo/bar.txt",
			expect:      "/tmp/foo/bar.txt",
		},
		{
			description: "ftp path",
			URL:         "ftp://localhost/tmp/foo/bar.txt",
			expect:      "/tmp/foo/bar.txt",
		},
		{
			description: "ftp root path",
			URL:         "ftp://localhost/",
			expect:      "/",
		},
		{
			description: "root path",
			URL:         "/",
			expect:      "/",
		},

		{
			description: "empty path",
			URL:         "",
			expect:      "/",
		},
		{
			description: "relative path",
			URL:         "abc/too.bar",
			expect:      "abc/too.bar",
		},
	}

	for _, useCase := range useCases {
		actual := Path(useCase.URL)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
