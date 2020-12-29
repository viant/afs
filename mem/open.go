package mem

import (
	"context"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
)

//Open downloads content for the supplied object
func (s *storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	root := s.Root
	file, err := root.File(location)
	if err != nil {
		return nil, err
	}
	generation := &option.Generation{}
	if _, ok := option.Assign(options, &generation); ok {
		generation.Generation = file.generation
	}
	return file.NewReader(), file.downloadError
}
