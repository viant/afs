package parrot

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/toolbox"
	"path"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {

	parent := toolbox.CallerDirectory(3)

	var useCases = []struct {
		description    string
		src            string
		dest           string
		useASCII       bool
		expectURL      string
		expectFragment string
	}{
		{
			description:    "folder mapping",
			src:            path.Join(parent, "test_data"),
			dest:           "mem://localhost/data",
			useASCII:       true,
			expectURL:      "mem://localhost/data/extract/txt.go",
			expectFragment: "var TXT = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit`)",
		},
	}

	fs := afs.New()
	for _, useCase := range useCases {
		ctx := context.Background()
		err := Generate(ctx, useCase.src, useCase.dest, useCase.useASCII)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := fs.DownloadWithURL(ctx, useCase.expectURL)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		if !assert.True(t, strings.Contains(string(actual), useCase.expectFragment), useCase.description) {
			fmt.Printf("%s\n", actual)
		}
	}

}
