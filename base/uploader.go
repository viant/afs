package base

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"os"
	"path"
	"strings"
	"sync/atomic"
)

type uploader struct {
	storage.Manager
}

//Close implements closer
func (u *uploader) Close() error {
	return nil
}

func (u *uploader) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	index := int32(0)
	handler := func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error {
		location := path.Join(parent, info.Name())

		if atomic.AddInt32(&index, 1) == 1 {
			if strings.HasSuffix(URL, location) {
				URL = string(URL[:len(URL)-len(location)])
			}
		}
		URL := url.Join(URL, location)
		if info.Mode()&os.ModeSymlink > 0 {
			if rawInfo, ok := info.(*file.Info); ok && rawInfo.Linkname != "" {
				options = append(options, rawInfo.Link)
			}
		}
		if info.IsDir() {
			return u.Manager.Create(ctx, URL, info.Mode(), info.IsDir(), options...)
		}
		return u.Manager.Upload(ctx, URL, info.Mode(), reader, options...)
	}
	return handler, u, nil
}

//NewUploader creates a new batch uploader
func NewUploader(manager storage.Manager) storage.BatchUploader {
	return &uploader{manager}
}
