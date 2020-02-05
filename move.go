package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

func (s *service) Move(ctx context.Context, sourceURL, destURL string, options ...storage.Option) error {
	sourceURL = url.Normalize(sourceURL, file.Scheme)
	destURL = url.Normalize(destURL, file.Scheme)
	destURL = s.updateDestURL(sourceURL, destURL)
	sourceOptions := option.NewSource()
	destOptions := option.NewDest()
	option.Assign(options, &sourceOptions, &destOptions)
	if url.IsSchemeEquals(sourceURL, destURL) {
		if sourceManager, err := s.manager(ctx, sourceURL, *sourceOptions); err == nil {
			if mover, ok := sourceManager.(storage.Mover); ok {
				if !s.IsAuthChanged(ctx, sourceManager, sourceURL, *destOptions) {
					return mover.Move(ctx, sourceURL, destURL, *sourceOptions...)
				}
			}
		}
	}
	_ = s.Delete(ctx, destURL, *destOptions...)
	if err := s.Copy(ctx, sourceURL, destURL, sourceOptions, destOptions); err != nil {
		return err
	}
	return s.Delete(ctx, sourceURL, *sourceOptions...)
}
