package option

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"testing"
)

func TestAppend(t *testing.T) {
	options := Append([]storage.Option{
		NewLocation("/tmp"),
		NewTimeout(100),
		NewSource(),
		NewDest(),
	}, NewBasicAuth("user", "pass"))
	assert.EqualValues(t, 5, len(options))
}
