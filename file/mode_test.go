package file

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewMode(t *testing.T) {

	var useCases = []struct {
		description string
		attrs       string
		expect      os.FileMode
		hasError    bool
	}{
		{
			description: "file rw",
			attrs:       "-rw-rw-rw-",
			expect:      0666,
		},
		{
			description: "invalid length",
			attrs:       "-rw-rw-r",
			hasError:    true,
		},
		{
			description: "file rwx",
			attrs:       "-rwxrw-rw-",
			expect:      0766,
		},
		{
			description: "directory rwx",
			attrs:       "drwxrw-rw-",
			expect:      0x800001F6,
		},
		{
			description: "directory rwx",
			attrs:       "drwxr-xr-x",
			expect:      DefaultDirOsMode,
		},
	}

	for _, useCase := range useCases {
		mode, err := NewMode(useCase.attrs)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, mode, useCase.expect, useCase.description)
	}

}
