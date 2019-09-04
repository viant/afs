package option

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPage(t *testing.T) {
	page := NewPage(2, 4)
	assert.True(t, page.ShallSkip())
	page.Increment()
	page.Increment()
	assert.False(t, page.ShallSkip())
	assert.False(t, page.HasReachedLimit())
	page.Increment()
	assert.False(t, page.HasReachedLimit())
	page.Increment()
	assert.True(t, page.HasReachedLimit())

}
