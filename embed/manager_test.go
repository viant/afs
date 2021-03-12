package embed_test

import (
	"context"
	"embed"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"strings"
	"testing"
)

//go:embed test/*
var eFs embed.FS

func TestManager_List(t *testing.T) {
	fs := afs.New()
	objects, err := fs.List(context.Background(), "embed:///test", eFs)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(objects))
	for _, object := range objects {
		if object.IsDir() {
			continue
		}
		data, err := fs.Download(context.Background(), object)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(string(data), "."))
	}

}
