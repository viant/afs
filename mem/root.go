package mem

import (
	"context"
	"github.com/viant/afs/url"
)

//Root returns memory system root folder for supplied base URL
func (s *manager) Root(ctx context.Context, baseURL string) *Folder {
	baseURL, _ = url.Base(baseURL, Scheme)
	srv, err := s.Storager(ctx, baseURL)
	if err != nil {
		return nil
	}
	memStorager, ok := srv.(*storager)
	if !ok {
		return nil
	}
	return memStorager.Root
}
