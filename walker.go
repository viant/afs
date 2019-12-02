package afs

import (
	"context"
	"errors"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"github.com/viant/afs/walker"
)

//Walk visits all location recursively within provided sourceURL
func (s *service) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	if URL == "" {
		return errors.New("URL was empty")
	}
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return err
	}
	URL = url.Normalize(URL, file.Scheme)
	managerWalker, ok := manager.(storage.Walker)
	if ok {
		return managerWalker.Walk(ctx, URL, handler, options...)
	}
	managerWalker = walker.New(manager)
	return managerWalker.Walk(ctx, URL, handler, options...)
}
