package option

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCrc(t *testing.T) {
	{
		crcHash := NewCrc([]byte("test is test"))
		actual := crcHash.Encode()
		assert.EqualValues(t, 0x84cd7d5, crcHash.Hash)
		assert.EqualValues(t, "CEzX1Q==", actual)
	}
	{
		crcHash := &Crc{}
		err := crcHash.Decode("CEzX1Q==")
		assert.Nil(t, err)
		assert.EqualValues(t, 0x84cd7d5, crcHash.Hash)
	}

}
