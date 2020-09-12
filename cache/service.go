package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"sync/atomic"
	"time"
)

type service struct {
	scheme   string
	host     string
	useCache int32
	baseURL  string
	cacheURL string
	next     *time.Time
	modified *time.Time
	afs.Service
}

func (s *service) canUseCache() bool {
	return atomic.LoadInt32(&s.useCache) == 1
}

func (s *service) Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return s.OpenURL(ctx, object.URL(), options...)
}

func (s *service) Object(ctx context.Context, URL string, options ...storage.Option) (storage.Object, error) {
	if e := s.reloadIfNeeded(ctx); e != nil {
		fmt.Printf("failed to reload %v", e)
		atomic.StoreInt32(&s.useCache, 0)
	}
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
	obj, _ := s.Object(ctx, URL, options...)
	return obj != nil, nil
}

func (s *service) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	if e := s.reloadIfNeeded(ctx); e != nil {
		fmt.Printf("failed to reload %v", e)
		atomic.StoreInt32(&s.useCache, 0)
	}
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
	if e := s.reloadIfNeeded(ctx); e != nil {
		fmt.Printf("failed to reload %v", e)
		atomic.StoreInt32(&s.useCache, 0)
	}
	if !s.canUseCache() {
		return s.Service.List(ctx, URL, options...)
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

func (s *service) reloadIfNeeded(ctx context.Context) error {
	if s.next != nil && s.next.After(time.Now()) {
		return nil
	}
	s.setNextRun(time.Now().Add(time.Second))
	cacheObject, _ := s.Service.Object(ctx, s.cacheURL)
	if cacheObject == nil {
		if e := s.build(ctx); e != nil {
			log.Printf("failed to build cache: %v %v", s.cacheURL, e)
		}
		cacheObject, _ = s.Service.Object(ctx, s.cacheURL)
		if cacheObject == nil {
			atomic.StoreInt32(&s.useCache, 0)
			return nil
		}
	}
	atomic.CompareAndSwapInt32(&s.useCache, 0, 1)
	if s.modified != nil && s.modified.Equal(cacheObject.ModTime()) {
		return nil
	}

	reader, err := s.Service.OpenURL(ctx, s.cacheURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()
	cache := &Cache{}
	if err = json.NewDecoder(reader).Decode(cache); err != nil {
		return err
	}
	for _, item := range cache.Items {
		URL := strings.Replace(item.URL, s.scheme, mem.Scheme, 1)
		if err = s.Service.Upload(ctx, URL, file.DefaultFileOsMode, bytes.NewReader(item.Data), item.ModTime); err != nil {
			break
		}
	}
	modTime := cacheObject.ModTime()
	s.modified = &modTime
	return err
}

func (s *service) build(ctx context.Context) error {
	objects, err := s.Service.List(ctx, s.baseURL, option.NewRecursive(true))
	if err != nil {
		return err
	}
	var items = make([]*Entry, 0)
	for _, obj := range objects {
		if obj.IsDir() || obj.Name() == CacheFile {
			continue
		}
		reader, err := s.Service.OpenURL(ctx, obj.URL())
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return err
		}
		items = append(items, &Entry{
			URL:     obj.URL(),
			Data:    data,
			Size:    obj.Size(),
			ModTime: obj.ModTime(),
		})
	}
	entries := &Cache{
		Items: items,
	}
	JSON, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	if exists, _ := s.Service.Exists(ctx, s.cacheURL); exists {
		return nil
	}
	err = s.Service.Upload(ctx, s.cacheURL, file.DefaultFileOsMode, bytes.NewReader(JSON), option.NewGeneration(true, 0))
	if isRateError(err) || isPreConditionError(err) { //ignore rate or generation errors
		err = nil
	}
	return err
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
func New(baseURL string, fs afs.Service) afs.Service {
	scheme := url.Scheme(baseURL, file.Scheme)
	if path.Ext(baseURL) != "" {
		baseURL, _ = url.Split(baseURL, scheme)
	}
	return &service{
		baseURL:  baseURL,
		host:     url.Host(baseURL),
		cacheURL: url.Join(baseURL, CacheFile),
		scheme:   scheme,
		Service:  fs,
	}
}
