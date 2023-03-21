package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io/ioutil"
	"strings"
	"sync"
)

func uploadCacheFile(ctx context.Context, cache *Cache, cacheURL string, service afs.Service) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	if strings.HasSuffix(cacheURL, ".gz") {
		data, _ = compressWithGzip(data)
	}
	err = service.Upload(ctx, cacheURL, file.DefaultFileOsMode, bytes.NewReader(data))
	if isRateError(err) || isPreConditionError(err) { //ignore rate or generation errors
		err = nil
	}
	return err
}

func build(ctx context.Context, baseURL, cacheName string, service afs.Service, opts ...storage.Option) (*Cache, error) {
	opts = append(opts, option.NewRecursive(true))
	objects, err := service.List(ctx, baseURL, opts...)
	if err != nil {
		return nil, err
	}
	var items = make([]*Entry, 0)
	entries := NewEntries(&items)
	wg := sync.WaitGroup{}
	for i, obj := range objects {
		if obj.IsDir() || obj.Name() == cacheName {
			continue
		}
		wg.Add(1)
		go func(object storage.Object) {
			defer wg.Done()
			reader, oErr := service.OpenURL(ctx, object.URL())
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
	return cacheEntries, nil
}
