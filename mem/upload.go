package mem

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"os"
)

//Upload writes fakeReader TestContent to supplied URL path.
func (s *storager) Upload(ctx context.Context, location string, mode os.FileMode, data []byte, options ...storage.Option) error {
	parent, err := s.parent(location, file.DefaultDirOsMode)
	if err != nil {
		return err
	}
	memFile := NewFile(location, mode, data)
	return parent.Put(memFile.Object)
}
