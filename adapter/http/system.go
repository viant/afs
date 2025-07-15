package http

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"net/http"
)

type Filesystem struct {
	fs      afs.Service
	dir     string
	options []storage.Option
}

func (f *Filesystem) Open(name string) (http.File, error) {
	object, err := f.fs.Object(context.Background(), url.Join(f.dir, name), f.options...)
	if err != nil {
		return nil, err
	}
	return NewFile(object, f.fs, f.options...)
}

// New creates http filesystem
func New(fs afs.Service, dir string, options ...storage.Option) http.FileSystem {
	return &Filesystem{fs: fs, dir: dir, options: options}
}
