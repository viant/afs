package http

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/url"
	"net/http"
)

type Filesystem struct {
	fs  afs.Service
	dir string
}

func (f *Filesystem) Open(name string) (http.File, error) {
	object, err := f.fs.Object(context.Background(), url.Join(f.dir, name))
	if err != nil {
		return nil, err
	}
	return NewFile(object, f.fs)
}

// New creates http filesystem
func New(fs afs.Service, dir string) http.FileSystem {
	return &Filesystem{fs: fs, dir: dir}
}
