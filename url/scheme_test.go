package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScheme(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
		expect      string
	}{

		{
			description: "raw path",
			URL:         "/foo/var.txt",
			expect:      "file",
		},
		{
			description: "fpt path",
			URL:         "ftp://localhosy/foo/var.txt",
			expect:      "ftp",
		},
	}

	for _, useCase := range useCases {
		actual := Scheme(useCase.URL, "file")
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
