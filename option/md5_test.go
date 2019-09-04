package option

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMd5(t *testing.T) {
	md5Hash := NewMd5([]byte("test is test"))
	hash := []byte{0x97, 0x6f, 0x7e, 0xa4, 0x4c, 0xce, 0x92, 0x2e, 0x6e, 0x5a, 0x27, 0x57, 0xa7, 0x87, 0x25, 0xe8}

	{
		actual := md5Hash.Encode()
		assert.EqualValues(t, hash, md5Hash.Hash)
		assert.EqualValues(t, "l29+pEzOki5uWidXp4cl6A==", actual)
	}
	{
		md5Hash := &Md5{}
		err := md5Hash.Decode("l29+pEzOki5uWidXp4cl6A==")
		assert.Nil(t, err)
		assert.EqualValues(t, hash, md5Hash.Hash)
	}

}
