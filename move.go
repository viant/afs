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
	sourceScheme := url.Scheme(sourceURL, file.Scheme)
	destScheme := url.Scheme(destURL, file.Scheme)

	sourceOptions := option.NewSource()
	destOptions := option.NewDest()
	var matcher option.WalkerMatcher
	_, _ = option.Assign(options, &sourceOptions, &destOptions, &matcher)
	if sourceScheme == destScheme {
		if manager, err := s.manager(ctx, sourceURL, *sourceOptions...); err == nil {
			if mover, ok := manager.(storage.Mover); ok {
				return mover.Move(ctx, sourceURL, destURL, options...)
			}
		}
	}
	_ = s.Delete(ctx, destURL, *destOptions...)
	if err := s.Copy(ctx, sourceURL, destURL, sourceOptions, destOptions); err != nil {
		return err
	}
	return s.Delete(ctx, sourceURL, *sourceOptions...)
}
