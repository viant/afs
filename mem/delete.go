package mem

import (
	"context"
	"github.com/viant/afs/storage"
	"path"
)

//Delete removes file or directory
func (s *storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	parent, err := s.parent(location, 0)
	if err != nil {
		return err
	}
	_, name := path.Split(location)
	return parent.Delete(name)
}
