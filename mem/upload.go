package mem

import (
	"context"
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"os"
	"time"
)

//Upload writes fakeReader TestContent to supplied URL path.
func (s *storager) Upload(ctx context.Context, location string, mode os.FileMode, data []byte, options ...storage.Option) error {
	parent, err := s.parent(location, file.DefaultDirOsMode)
	if err != nil {
		return err
	}
	modTime := time.Now()
	option.Assign(options, &modTime)
	memFile := NewFile(location, mode, data, modTime)
	return parent.Put(memFile.Object)
}
