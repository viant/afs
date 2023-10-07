package http

import (
	"context"
	"github.com/viant/afs"
	"net/http"
)

type Filesystem struct {
	fs afs.Service
}

func (f *Filesystem) Open(name string) (http.File, error) {
	object, err := f.fs.Object(context.Background(), name)
	if err != nil {
		return nil, err
	}
	return NewFile(object, f.fs)
}

// New creates http filesystem
func New(fs afs.Service) http.FileSystem {
	return &Filesystem{fs: fs}
}
