package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

func (s *service) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	URL = url.Normalize(URL, file.Scheme)
	recursive := &option.Recursive{}
	options, _ = option.Assign(options, &recursive)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return nil, err
	}
	var result = make([]storage.Object, 0)
	return result, list(ctx, manager, URL, recursive.Flag, options, &result)
}

func list(ctx context.Context, lister storage.Lister, URL string, recursive bool, options []storage.Option, result *[]storage.Object) error {
	objects, err := lister.List(ctx, URL, options...)
	if err != nil {
		return err
	}

	dirs := make([]storage.Object, 0)
	for i, object := range objects {
		if object.IsDir() && recursive {
			if !url.Equals(URL, object.URL()) {
				dirs = append(dirs, objects[i])
			}
			continue
		}
		*result = append(*result, objects[i])
	}

	if recursive {
		var matchOpt option.Match
		var matcherOpt option.Matcher
		if _, has := option.Assign(options, &matcherOpt, &matchOpt); has {
			dirMatcher := &matcher.Basic{Directory: &recursive}
			dirs, err = lister.List(ctx, URL, dirMatcher.Match)
			if err != nil {
				return err
			}
		}
		for i := 0; i < len(dirs); i++ {
			if url.Equals(URL, dirs[i].URL()) {
				continue
			}
			*result = append(*result, dirs[i])
			if err = list(ctx, lister, dirs[i].URL(), recursive, options, result); err != nil {
				return err
			}
		}
	}
	return nil
}
