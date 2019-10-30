package base

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/option"
	"io"
	"testing"
)

func TestStreamReader_Read(t *testing.T) {

	var useCases = []struct {
		description string
		input       string
		parts       []string
		partSize    int
		readBuffer  []byte
	}{

		{
			description: "dest readBuffer larger than stream readBuffer",
			partSize:    15,
			readBuffer:  make([]byte, 20),
			input:       "0123456789012345678901234567890123456789",
			parts: []string{
				"01234567890123456789", "01234567890123456789",
			},
		},

		{
			description: "dest readBuffer larger than stream readBuffer",
			partSize:    3,
			readBuffer:  make([]byte, 4),
			input:       "abcdefghij",
			parts: []string{
				"abcd", "efgh", "ij",
			},
		},
		{
			description: "dest readBuffer smaller than stream readBuffer",
			partSize:    5,
			readBuffer:  make([]byte, 3),
			input:       "abcdefghij",
			parts: []string{
				"abc", "def", "ghi", "j",
			},
		},
		{
			description: "dest readBuffer and stream buffer the same",
			partSize:    5,
			readBuffer:  make([]byte, 5),
			input:       "abcdefghij",
			parts: []string{
				"abcde", "fghij",
			},
		},
		{
			description: "dest readBuffer and stream buffer the same with overflow",
			partSize:    5,
			readBuffer:  make([]byte, 5),
			input:       "abcdefghijk",
			parts: []string{
				"abcde", "fghij", "k",
			},
		},
		{
			description: "dest readBuffer and stream buffer the same with overflow 4",
			partSize:    5,
			readBuffer:  make([]byte, 5),
			input:       "abcdefghijklmn",
			parts: []string{
				"abcde", "fghij", "klmn",
			},
		},
		{
			description: "dest readBuffer and stream buffer the same with overflow 4",
			partSize:    8,
			readBuffer:  make([]byte, 10),
			input:       "abcdefghijklmnoprstuvxyz",
			parts: []string{
				"abcdefghij", "klmnoprstu", "vxyz",
			},
		},
	}

	for _, useCase := range useCases {

		reader := NewStreamReader(option.NewStream(useCase.partSize, len(useCase.input)), &testRanger{text: useCase.input})
		actual := make([]string, 0)
		for i := range useCase.parts {
			read, err := reader.Read(useCase.readBuffer)
			if !assert.Nil(t, err, fmt.Sprintf(useCase.description+" / %v", i)) {
				continue
			}
			if !assert.True(t, read > 0) {
				continue
			}
			text := string(useCase.readBuffer[:read])
			actual = append(actual, text)
		}
		read, err := reader.Read(useCase.readBuffer)
		assert.Equal(t, io.EOF, err, useCase.description)
		assert.Equal(t, 0, read, useCase.description)
		assert.EqualValues(t, useCase.parts, actual)
	}

}

type testRanger struct {
	text string
	from int64
}

func (t *testRanger) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return 0, fmt.Errorf("whence usupported: %v", whence)
	}
	t.from = offset
	return 0, nil
}

func (t *testRanger) Read(dest []byte) (int, error) {
	to := int(t.from) + len(dest)
	copy(dest, t.text[t.from:to])
	return (to - int(t.from)), nil
}
