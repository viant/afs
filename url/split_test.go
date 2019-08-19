package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit(t *testing.T) {

	var useCases = []struct {
		description string
		URL         string
		URLParent   string
		URIName     string
	}{
		{
			description: "file uri",
			URL:         "scp://localhost/foo/bar.txt",
			URLParent:   "scp://localhost/foo",
			URIName:     "bar.txt",
		},
		{
			description: "path uri",
			URL:         "scp://localhost/foo/bar",
			URLParent:   "scp://localhost/foo",
			URIName:     "bar",
		},
		{
			description: "path uri",
			URL:         "scp://localhost/foo/bar/",
			URLParent:   "scp://localhost/foo",
			URIName:     "bar",
		},
	}

	for _, useCase := range useCases {
		parent, name := Split(useCase.URL, "file")
		assert.Equal(t, useCase.URLParent, parent, useCase.description)
		assert.Equal(t, useCase.URIName, name, useCase.description)
	}

}
