package base

import (
	"context"
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"os"
	"path"
)

type uploader struct {
	storage.Manager
}

func (u *uploader) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	handler := func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error {
		location := path.Join(parent, info.Name())
		URL := url.Join(URL, location)
		if info.Mode()&os.ModeSymlink > 0 {
			if rawInfo, ok := info.(*file.Info); ok && rawInfo.Linkname != "" {
				fmt.Printf("is link %v\n", rawInfo)
				options = append(options, rawInfo.Link)
			}
		}
		if info.IsDir() {
			return u.Manager.Create(ctx, URL, info.Mode(), info.IsDir(), options...)
		}
		return u.Manager.Upload(ctx, URL, info.Mode(), reader, options...)
	}
	return handler, u.Manager, nil
}

//NewUploader creates a new batch uploader
func NewUploader(manager storage.Manager) storage.BatchUploader {
	return &uploader{manager}
}
