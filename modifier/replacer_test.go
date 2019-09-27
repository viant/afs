package modifier

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/file"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestReplace(t *testing.T) {
	replaceer := Replace(map[string]string{
		"test": "Test",
	})

	info := file.NewInfo("blah", 0, 0644, time.Now(), false)
	_, reader, err := replaceer(info, ioutil.NopCloser(strings.NewReader("test is test")))
	assert.Nil(t, err)
	actual, _ := ioutil.ReadAll(reader)
	assert.EqualValues(t, "Test is Test", actual)
}
