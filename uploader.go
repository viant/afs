package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"os"
	"path"
)

//UploadInBatch default implementation for UploadInBatch
func (s *service) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return nil, nil, err
	}
	batchUploader, ok := manager.(storage.BatchUploader)
	if ok {
		return batchUploader.Uploader(ctx, URL, options...)
	}
	handler := func(ctx context.Context, relativePath string, info os.FileInfo, reader io.Reader) error {
		location := path.Join(relativePath, info.Name())
		URL := url.Join(URL, location)
		if info.IsDir() {
			return manager.Create(ctx, URL, info.Mode(), info.IsDir(), options)
		}
		return manager.Upload(ctx, URL, info.Mode(), reader, options...)
	}
	return handler, manager, nil
}
