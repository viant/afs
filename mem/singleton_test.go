package mem

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleton(t *testing.T) {

	s1 := Singleton()
	s2 := Singleton()
	assert.Equal(t, s1, s2)
	ResetSingleton()
	s3 := Singleton()
	assert.True(t, s3 != s1)
}
