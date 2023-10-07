package http

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/storage"
	"io"
	"io/fs"
	"net/http"
)

type file struct {
	fs     afs.Service
	reader io.ReadCloser
	object storage.Object
}

func (f *file) Close() error {
	if f.reader == nil {
		return fmt.Errorf("reader was nil")
	}
	return f.reader.Close()
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.reader == nil {
		return 0, fmt.Errorf("reader was nil")
	}
	return f.reader.Read(p)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := f.reader.(io.Seeker)
	if !ok {
		return 0, fmt.Errorf("invalid reader type: %T", seeker)
	}
	return seeker.Seek(offset, whence)
}

func (f *file) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.object.IsDir() {
		return nil, fmt.Errorf("not directory: %v ", f.object.URL())
	}
	objects, err := f.fs.List(context.Background(), f.object.URL())
	if err != nil {
		return nil, err
	}
	var result = make([]fs.FileInfo, 0, len(objects))
	for i := range objects {
		result = append(result, objects[i])
	}
	if count > 0 {
		result = result[:count]
	}
	return result, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.object, nil
}

// NewFile creates a http.File
func NewFile(object storage.Object, fs afs.Service) (http.File, error) {
	ret := &file{object: object, fs: fs}
	if object.IsDir() {
		return ret, nil
	}
	reader, err := fs.Open(context.Background(), object)
	if err != nil {
		return nil, err
	}
	ret.reader = reader
	return ret, nil
}
