package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBase(t *testing.T) {
	var useCases = []struct {
		description   string
		URL           string
		expectBaseURL string
		expectPath    string
	}{
		{
			description:   "simple path",
			URL:           "/tmp/abc.txt",
			expectBaseURL: "file://localhost",
			expectPath:    "/tmp/abc.txt",
		},
		{
			description:   "scp path",
			URL:           "scp://10.1.22.1/tmp/abc.txt",
			expectBaseURL: "scp://10.1.22.1",
			expectPath:    "/tmp/abc.txt",
		},

		{
			description:   "ftp path",
			URL:           "ftp://10.1.22.1:33/tmp/abc.txt",
			expectBaseURL: "ftp://10.1.22.1:33",
			expectPath:    "/tmp/abc.txt",
		},

		{
			description:   "ftp root",
			URL:           "ftp://10.1.22.1:33/",
			expectBaseURL: "ftp://10.1.22.1:33",
			expectPath:    "/",
		},
		{
			description:   "mem",
			URL:           "mem:///var/folders/gl/5550g3kj6tn1rbz8chqx1c61ycmmm1/",
			expectBaseURL: "mem://",
			expectPath:    "/var/folders/gl/5550g3kj6tn1rbz8chqx1c61ycmmm1/",
		},
	}

	for _, useCase := range useCases {
		actualBaseURL, actualPath := Base(useCase.URL, "file")
		assert.EqualValues(t, useCase.expectBaseURL, actualBaseURL, useCase.description)
		assert.EqualValues(t, useCase.expectPath, actualPath, useCase.description)

	}
}
