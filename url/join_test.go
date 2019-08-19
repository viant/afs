package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoin(t *testing.T) {

	var useCases = []struct {
		description string
		baseURL     string
		elemenets   []string
		expect      string
	}{
		{
			description: "relative elements",
			baseURL:     "ftp://localhost",
			elemenets:   []string{"foo", "bar.txt"},
			expect:      "ftp://localhost/foo/bar.txt",
		},
		{
			description: "trimmed elements",
			baseURL:     "ftp://localhost",
			elemenets:   []string{"/foo/", "/bar.txt"},
			expect:      "ftp://localhost/foo/bar.txt",
		},
		{
			description: "base path only",
			baseURL:     "ftp://localhost",
			elemenets:   []string{},
			expect:      "ftp://localhost",
		},
	}

	for _, useCase := range useCases {
		actual := Join(useCase.baseURL, useCase.elemenets...)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
