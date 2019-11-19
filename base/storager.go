package base

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"os"
)

type List func(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error)

//Storager represents a base storager
type Storager struct {
	List func(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error)
}

//Get returns an object for supplied location
func (s *Storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	options = append(options, option.NewPage(0, 1))
	objects, err := s.List(ctx, location, options)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, errors.Errorf("failed to get object: %v", location)
	}
	return objects[0], nil
}
