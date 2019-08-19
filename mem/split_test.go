package mem

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitPath(t *testing.T) {

	var useCases = []struct {
		description string
		URLPath     string
		expect      []string
	}{
		{
			description: "root path",
			URLPath:     "/",
			expect:      []string{},
		},
		{
			description: "folder path",
			URLPath:     "/folder",
			expect:      []string{"folder"},
		},
		{
			description: "folder path with slash",
			URLPath:     "/folder/",
			expect:      []string{"folder"},
		},
		{
			description: "subfolder",
			URLPath:     "/folder/subfolder",
			expect:      []string{"folder", "subfolder"},
		},
		{
			description: "file ",
			URLPath:     "/folder/subfolder/file.txt",
			expect:      []string{"folder", "subfolder", "file.txt"},
		},
	}

	for _, useCase := range useCases {

		actual := SplitPath(useCase.URLPath)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)

	}

}
