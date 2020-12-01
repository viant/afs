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

func TestGenerateWithMem(t *testing.T) {

	parent := toolbox.CallerDirectory(3)

	var useCases = []struct {
		description    string
		src            string
		dest           string
		useASCII       bool
		expectFragment string
	}{
		{
			description:    "folder mapping",
			src:            path.Join(parent, "test"),
			dest:           "mem://localhost/gen/test1.go",
			useASCII:       true,
			expectFragment: "func run() {}",
		},
		{
			description:    "file mapping",
			src:            path.Join(parent, "test/runner.go"),
			dest:           "mem://localhost/gen/test2.go",
			useASCII:       true,
			expectFragment: "func run() {}",
		},
	}

	fs := afs.New()
	for _, useCase := range useCases {
		ctx := context.Background()
		err := GenerateWithMem(ctx, useCase.src, useCase.dest, useCase.useASCII)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual, err := fs.DownloadWithURL(ctx, useCase.dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		if !assert.True(t, strings.Contains(string(actual), useCase.expectFragment), useCase.description) {
			fmt.Printf("%s\n", actual)
		}
	}

}
