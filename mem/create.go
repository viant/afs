package mem

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"os"
)

//Create creates a new file or directory
func (s *storager) Create(ctx context.Context, location string, mode os.FileMode, reader io.Reader, isDir bool, options ...storage.Option) error {
	root := s.Root
	if isDir {
		_, err := root.Folder(location, mode)
		return err
	}
	return s.Upload(ctx, location, mode, reader)
}
