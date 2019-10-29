package mem

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"os"
	"time"
)

//Upload writes fakeReader TestContent to supplied URL path.
func (s *storager) Upload(ctx context.Context, location string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	parent, err := s.parent(location, file.DefaultDirOsMode)
	if err != nil {
		return err
	}
	var data []byte
	if reader != nil {
		if data, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	}
	modTime := time.Now()
	option.Assign(options, &modTime)
	memFile := NewFile(location, mode, data, modTime)
	return parent.Put(memFile.Object)
}
