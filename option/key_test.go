package option

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAES256Key(t *testing.T) {

	{
		data := []byte("this is test    this is test key")
		key, err := NewAES256Key(data)
		assert.Nil(t, err)
		assert.Equal(t, key.Key, data)
		err = key.Init()
		assert.Nil(t, err)
		err = key.Validate()
		assert.Nil(t, err)
	}
	{
		//invalid lenth
		_, err := NewAES256Key([]byte("abc"))
		assert.NotNil(t, err)
	}

}
