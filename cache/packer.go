package cache

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"path"
	"strings"
)

//Package creates cache file for source URL with rewrite
func Package(ctx context.Context, sourceURL string, rewriteBaseURL string, options ...storage.Option) error {
	var cacheOption = &option.Cache{}
	option.Assign(options, &cacheOption)
	if cacheOption.Name == "" {
		cacheOption.Name = CacheFile
	}
	cacheOption.Init()
	cacheURL := path.Join(rewriteBaseURL, cacheOption.Name)
	fs := afs.New()
	cache, err := build(ctx, sourceURL, cacheOption.Name, fs, options...)
	if err != nil || len(cache.Items) == 0 {
		return err
	}
	sourceURL = url.Normalize(sourceURL, file.Scheme)
	for _, entry := range cache.Items {
		location := strings.Replace(entry.URL, sourceURL, "", 1)
		entry.URL = url.Join(rewriteBaseURL, location)
	}
	return uploadCacheFile(ctx, cache, cacheURL, fs)
}
