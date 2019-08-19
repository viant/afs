package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHost(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
		expect      string
	}{

		{
			URL:    "/tmp",
			expect: Localhost,
		},
		{
			URL:    "scp://myhost:22/dad",
			expect: "myhost:22",
		},

		{
			description: "host only",
			URL:         "scp://myhost:22",
			expect:      "myhost:22",
		},
	}

	for _, useCase := range useCases {
		actual := Host(useCase.URL)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
