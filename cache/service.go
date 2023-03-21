package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type service struct {
	cacheName string
	scheme    string
	host      string
	useCache  int32
	baseURL   string
	cacheURL  string
	next      *time.Time
	modified  *time.Time
	exclusion *matcher.Ignore
	at        time.Time
	checksum  int
	refresh   *option.RefreshInterval
	afs.Service
	logger *option.Logger
}

func (s *service) canUseCache() bool {
	return atomic.LoadInt32(&s.useCache) == 1
}

func (s *service) Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	if s.isExcluded(object.URL(), object) {
		return s.Service.Open(ctx, object, options...)
	}
	return s.OpenURL(ctx, object.URL(), options...)
}

func (s *service) Object(ctx context.Context, URL string, options ...storage.Option) (storage.Object, error) {
	if s.isExcluded(URL, nil) {
		return s.Service.Object(ctx, URL, options...)
	}
	s.reloadIfNeeded(ctx)
	if !s.canUseCache() {
		return s.Service.Object(ctx, URL, options...)
	}
	cacheURL := strings.Replace(URL, s.scheme, mem.Scheme, 1)
	obj, _ := s.Service.Object(ctx, cacheURL, options...)
	if obj != nil {
		return s.rewriteObject(obj), nil
	}
	return s.Service.Object(ctx, URL, options...)
}

func (s *service) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {
	if s.isExcluded(URL, nil) {
		return s.Service.Exists(ctx, URL, options...)
	}
	obj, _ := s.Object(ctx, URL, options...)
	return obj != nil, nil
}

func (s *service) isExcluded(candidateURL string, info os.FileInfo) bool {
	if s.exclusion == nil {
		return false
	}
	if index := strings.Index(s.baseURL, candidateURL); index != -1 {
		candidateURL = candidateURL[index+len(s.baseURL):]
	}
	parent, name := path.Split(candidateURL)
	if info == nil { //default info
		info = file.NewInfo(name, 1, file.DefaultFileOsMode, s.at, false)
	}
	if !s.exclusion.Match(parent, info) {
		return true
	}
	return false
}

func (s *service) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	s.reloadIfNeeded(ctx)
	if !s.canUseCache() {
		return s.Service.OpenURL(ctx, URL, options...)
	}
	cacheURL := strings.Replace(URL, s.scheme, mem.Scheme, 1)
	reader, err := s.Service.OpenURL(ctx, cacheURL, options...)
	if err == nil {
		return reader, err
	}
	return s.Service.OpenURL(ctx, URL, options...)
}

func (s *service) rewriteObject(obj storage.Object) storage.Object {
	URL := strings.Replace(obj.URL(), mem.Scheme, s.scheme, 1)
	URL = strings.Replace(URL, url.Localhost, s.host, 1)
	return object.New(URL, obj, obj.Sys())
}

func (s *service) rewriteObjects(objects []storage.Object) []storage.Object {
	var result = make([]storage.Object, 0)
	for i := range objects {
		result = append(result, s.rewriteObject(objects[i]))
	}
	return result
}

func (s *service) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	s.reloadIfNeeded(ctx)
	if !s.canUseCache() {
		return s.Service.List(ctx, URL, options...)
	}
	if s.exclusion != nil {
		options = append(options, s.exclusion)
	}
	cacheURL := strings.Replace(URL, s.scheme, mem.Scheme, 1)
	if objects, _ := s.Service.List(ctx, cacheURL, options...); len(objects) > 0 {
		return s.rewriteObjects(objects), nil
	}
	return s.Service.List(ctx, URL, options...)
}

func (s *service) setNextRun(next time.Time) {
	s.next = &next
}

func (s *service) reloadIfNeeded(ctx context.Context) {
	if s.next != nil && s.next.After(time.Now()) {
		return
	}
	started := time.Now()
	defer func() {
		s.logger.Logf("rebuild cache %v after %s\n", s.cacheURL, time.Since(started))
	}()
	s.setNextRun(time.Now().Add(s.refresh.Duration()))
	err := s.reloadCache(ctx)
	if err != nil {
		fmt.Printf("failed to reload cache: %v", err)
		atomic.StoreInt32(&s.useCache, 0)
	} else {
		atomic.CompareAndSwapInt32(&s.useCache, 0, 1)
	}

}

func (s *service) reloadCache(ctx context.Context) error {
	cacheObject, _ := s.Service.Object(ctx, s.cacheURL, option.NewObjectKind(true))
	var cache *Cache
	var err error
	if s.shallRebuildCache(cacheObject) {
		if cache, err = s.build(ctx); err != nil {
			log.Printf("failed to build cache: %v %v", s.cacheURL, err)
		}
		if err = s.uploadCache(ctx, cache, cacheObject); err != nil {
			return err
		}
	}
	if cache == nil {
		data, err := s.Service.DownloadWithURL(ctx, s.cacheURL)
		if err = json.Unmarshal(data, cache); err != nil {
			return err
		}
	}
	s.syncCache(ctx, cache)
	return err
}

func (s *service) syncCache(ctx context.Context, cache *Cache) {
	var err error
	for _, item := range cache.Items {
		URL := strings.Replace(item.URL, s.scheme, mem.Scheme, 1)
		if err = s.Service.Upload(ctx, URL, file.DefaultFileOsMode, bytes.NewReader(item.Data), item.ModTime); err != nil {
			break
		}
	}
}

func (s *service) build(ctx context.Context) (*Cache, error) {
	var opts []storage.Option
	opts = append(opts, option.NewRecursive(true))
	if s.exclusion != nil {
		opts = append(opts, s.exclusion)
	}
	objects, err := s.Service.List(ctx, s.baseURL, opts...)
	if err != nil {
		return nil, err
	}

	var items = make([]*Entry, 0)
	entries := NewEntries(&items)
	wg := sync.WaitGroup{}
	for i, obj := range objects {
		if obj.IsDir() || obj.Name() == s.cacheName {
			continue
		}
		wg.Add(1)
		go func(object storage.Object) {
			defer wg.Done()
			reader, oErr := s.Service.OpenURL(ctx, object.URL())
			if err != nil {
				err = oErr
				return
			}
			data, oErr := ioutil.ReadAll(reader)
			_ = reader.Close()
			if oErr != nil {
				err = oErr
				return
			}
			entries.Append(&Entry{
				URL:     object.URL(),
				Data:    data,
				Size:    object.Size(),
				ModTime: object.ModTime(),
			})

		}(objects[i])

	}
	wg.Wait()
	cacheEntries := &Cache{
		Items: items,
	}
	JSON, err := json.Marshal(cacheEntries)
	if err != nil {
		return nil, err
	}
	err = s.Service.Upload(ctx, s.cacheURL, file.DefaultFileOsMode, bytes.NewReader(JSON))
	if isRateError(err) || isPreConditionError(err) { //ignore rate or generation errors
		err = nil
	}
	return cacheEntries, err
}

func (s *service) shallRebuildCache(cacheObject storage.Object) bool {
	if s.modified == nil || cacheObject == nil {
		return true
	}
	return s.modified.Equal(cacheObject.ModTime())
}

func (s *service) uploadCache(ctx context.Context, cache *Cache, prev storage.Object) error {
	cacheObject, _ := s.Service.Object(ctx, s.cacheURL, option.NewObjectKind(true))
	if cacheObject == nil || (prev != nil && cacheObject.ModTime().Equal(prev.ModTime())) {
		data, err := json.Marshal(cache)
		if err != nil {
			return err
		}
		if err = s.Service.Upload(ctx, s.cacheURL, file.DefaultFileOsMode, bytes.NewReader(data)); err != nil {
			return err
		}
		if latest, _ := s.Service.Object(ctx, s.cacheURL, option.NewObjectKind(true)); latest != nil {
			mod := latest.ModTime()
			s.modified = &mod
		}
	}
	return nil
}

func isRateError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("%v", http.StatusTooManyRequests))
}

func isPreConditionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("%v", http.StatusPreconditionFailed))
}

//New create a cache service for supplied base URL
func New(baseURL string, fs afs.Service, opts ...storage.Option) afs.Service {
	var cacheName = &option.CacheName{}
	logger := &option.Logger{}
	option.Assign(opts, &cacheName, &logger)
	if cacheName.Name == "" {
		cacheName.Name = CacheFile
	}

	scheme := url.Scheme(baseURL, file.Scheme)
	if path.Ext(baseURL) != "" {
		baseURL, _ = url.Split(baseURL, scheme)
	}
	ret := &service{
		at:        time.Now(),
		cacheName: cacheName.Name,
		baseURL:   baseURL,
		host:      url.Host(baseURL),
		cacheURL:  url.Join(baseURL, cacheName.Name),
		scheme:    scheme,
		Service:   fs,
		refresh:   &option.RefreshInterval{},
		logger:    logger,
	}

	var ignore = &matcher.Ignore{}
	if _, ok := option.Assign(opts, &ignore); ok {
		ret.exclusion = ignore
	}
	option.Assign(opts, &ret.refresh)
	if ret.refresh.IntervalMs == 0 {
		ret.refresh.IntervalMs = 3000
	}
	return ret
}
