package mem

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
)

//Download downloads content for the supplied object
func (s *storager) Download(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	root := s.Root
	file, err := root.File(location)
	if err != nil {
		return nil, err
	}
	return file.NewReader(), file.downloadError
}
