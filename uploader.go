package afs

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
			return manager.Create(ctx, URL, info.Mode(), info.IsDir(), options...)
		}
		return manager.Upload(ctx, URL, info.Mode(), reader, options...)
	}
	return handler, manager, nil
}
