package embed

import (
	"embed"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

//go:embed test/*
var testEmbedFs embed.FS

func TestHolder_Add(t *testing.T) {

	embedFs := NewHolder()
	embedFs.Add("zerba.txt", "this is foo context")
	embedFs.Add("foo/bar.txt", "this is context")
	embedFs.Add("foo/dummy.txt", "this is context")
	embedFs.Add("foo/sub/dummy.txt", "this is context")
	fs := embedFs.EmbedFs()
	entries, err := fs.ReadDir("foo")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(entries))

	{
		fh, err := fs.Open("foo/bar.txt")
		assert.Nil(t, err)
		data, err := io.ReadAll(fh)
		assert.Nil(t, err)
		assert.EqualValues(t, "this is context", string(data))
		_ = fh.Close()
	}

	aHolder := NewHolder()
	aHolder.AddFs(&testEmbedFs, "embed:///test")
	merged := aHolder.EmbedFs()
	fmt.Println(merged)
}
