package scp

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"os"
	"testing"
	"time"
)

func TestNewInfo(t *testing.T) {

	var useCases = []struct {
		description    string
		createResponse string
		expectName     string
		expectMode     os.FileMode
		expectSize     int
		isDir          bool
		hasError       bool
	}{
		{
			description:    "create location",
			createResponse: "C0644 12 foo.bar\n",
			expectName:     "foo.bar",
			expectMode:     os.FileMode(0644),
			isDir:          false,
			expectSize:     12,
		},
		{
			description:    "create folder",
			createResponse: "D0755 0 test\n",
			expectName:     "test",
			expectMode:     os.FileMode(0755),
			isDir:          true,
			expectSize:     0,
		},
		{
			description:    "error response folder",
			createResponse: "D0755 0\n",
			hasError:       true,
		},
	}
	for _, useCase := range useCases {
		info, err := NewInfo(useCase.createResponse, nil)

		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {

		}
		assert.EqualValues(t, useCase.isDir, info.IsDir(), useCase.description)
		assert.EqualValues(t, useCase.expectMode, info.Mode(), useCase.description)
		assert.EqualValues(t, useCase.expectName, info.Name(), useCase.description)
		assert.EqualValues(t, useCase.expectSize, info.Size(), useCase.description)

	}

}

func TestInfoToCreateCmd(t *testing.T) {
	var useCases = []struct {
		description string
		info        os.FileInfo
		expect      string
	}{
		{
			description: "create location",
			info:        file.NewInfo("test.txt", 14, 0644, time.Now(), false),
			expect:      "C0644 14 test.txt\n",
		},
		{
			description: "create dir",
			info:        file.NewInfo("folder", 14, 0644, time.Now(), true),
			expect:      "D0644 0 folder\n",
		},
	}

	for _, useCase := range useCases {
		actual := InfoToCreateCmd(useCase.info)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
