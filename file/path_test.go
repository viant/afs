package file

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestPath(t *testing.T) {
	tempDir := os.TempDir()

	var useCases = []struct {
		description string
		location    string
		expect      string
	}{
		{
			description: "tmp folder path",
			location:    tempDir,
			expect:      strings.ReplaceAll(tempDir, `\`, `/`),
		},
	}

	for _, useCase := range useCases {
		actual := Path(useCase.location)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
