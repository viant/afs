package url

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScheme(t *testing.T) {

	var useCases = []struct {
		description  string
		URL          string
		expect       string
		extensionURL string
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
		{
			description:  "zip extended path",
			URL:          "s3:myBucket/root/path/app.zip/zip://localhost/foo/var.txt",
			expect:       "zip",
			extensionURL: "s3://myBucket/root/path/app.zip",
		},
	}

	for _, useCase := range useCases {
		actual := Scheme(useCase.URL, "file")
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
		schmeExtensionURL := SchemeExtensionURL(useCase.URL)
		assert.EqualValues(t, useCase.extensionURL, schmeExtensionURL, useCase.description)

	}

}
