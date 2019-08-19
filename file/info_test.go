package file

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNewInfo(t *testing.T) {

	var useCases = []struct {
		description      string
		name             string
		size             int64
		mode             os.FileMode
		modificationTime time.Time
		isDir            bool
	}{
		{
			description:      "folder test",
			name:             "abc",
			size:             2,
			mode:             0777,
			modificationTime: time.Now(),
			isDir:            false,
		},
		{
			description:      "dir test",
			name:             "abc",
			size:             2,
			mode:             0777,
			modificationTime: time.Now(),
			isDir:            true,
		},
	}

	for _, useCase := range useCases {
		info := NewInfo(useCase.name, useCase.size, useCase.mode, useCase.modificationTime, useCase.isDir)
		assert.EqualValues(t, useCase.name, info.Name())
		assert.EqualValues(t, useCase.size, info.Size())
		assert.EqualValues(t, useCase.mode, info.Mode())
		assert.EqualValues(t, useCase.modificationTime, info.ModTime())
		assert.EqualValues(t, useCase.isDir, info.IsDir())

	}

}
