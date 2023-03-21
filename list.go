package afs

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"sync"
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
	objects := storage.NewObjects(&result)
	return result, list(ctx, manager, URL, recursive.Flag, options, objects)
}

func list(ctx context.Context, lister storage.Lister, URL string, recursive bool, options []storage.Option, result *storage.Objects) error {
	objects, err := lister.List(ctx, URL, options...)
	if err != nil {
		return err
	}
	var matchFn option.Match
	var aMatcher option.Matcher

	_, hasMatchFn := option.Assign(options, &matchFn)
	_, hasMatcher := option.Assign(options, &aMatcher)

	dirs := make([]storage.Object, 0)
	for i, object := range objects {
		objectURL := object.URL()
		if !hasMatch(objectURL, hasMatchFn, matchFn, object, hasMatcher, aMatcher) {
			continue
		}
		if object.IsDir() && recursive {
			if !url.Equals(URL, object.URL()) {
				dirs = append(dirs, objects[i])
			}
			continue
		}
		result.Append(objects[i])
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
		wg := &sync.WaitGroup{}
		for i := 0; i < len(dirs); i++ {
			if i == 0 && url.Equals(URL, dirs[i].URL()) {
				continue
			}
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				result.Append(dirs[index])
				if lErr := list(ctx, lister, dirs[index].URL(), recursive, options, result); lErr != nil {
					err = lErr
				}
			}(i)
		}
		wg.Wait()
	}
	return err
}

func hasMatch(objectURL string, hasMatchFn bool, matchFn option.Match, object storage.Object, hasMatcher bool, aMatcher option.Matcher) bool {
	if !(hasMatcher && hasMatchFn) {
		return true
	}
	location := url.Path(objectURL)
	parent := url.Dir(location)
	if hasMatchFn && !matchFn(parent, object) {
		return false
	}
	if hasMatcher && !aMatcher.Match(parent, object) {
		return false
	}
	return true
}
