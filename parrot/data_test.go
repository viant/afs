package parrot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestData_AsBytesLiteral(t *testing.T) {

	var useCases = []struct {
		description string
		data        Data
		expect      string
		ASCII       bool
	}{
		{
			description: "short binary data",
			data:        []byte("this is test"),
			expect:      `[]byte{0x74,0x68,0x69,0x73,0x20,0x69,0x73,0x20,0x74,0x65,0x73,0x74}`,
		},
		{
			description: "long binary data",
			data:        []byte("this is test with extra data Lorem Ipsum"),
			expect: `[]byte{0x74,0x68,0x69,0x73,0x20,0x69,0x73,0x20,0x74,0x65,0x73,0x74,0x20,0x77,0x69,0x74,
0x68,0x20,0x65,0x78,0x74,0x72,0x61,0x20,0x64,0x61,0x74,0x61,0x20,0x4c,0x6f,0x72,
0x65,0x6d,0x20,0x49,0x70,0x73,0x75,0x6d}`,
		},
		{
			description: "literal  data",
			data:        []byte("this is test"),
			expect:      "[]byte(`this is test`)",
			ASCII:       true,
		},
	}

	for _, useCase := range useCases {
		actual := useCase.data.AsBytesLiteral(useCase.ASCII)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
