package option

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func TestReplace(t *testing.T) {
	replaceer := Replace(map[string]string{
		"test": "Test",
	})
	reader, err := replaceer(nil, strings.NewReader("test is test"))
	assert.Nil(t, err)
	actual, _ := ioutil.ReadAll(reader)
	assert.EqualValues(t, "Test is Test", actual)
}
