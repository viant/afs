package afs

import (
	"context"
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

//Service represents storage storage
type Service interface {
	storage.Lister
	storage.Opener
	storage.Uploader
	storage.BatchUploader
	storage.Deleter
	storage.Creator
	storage.Walker
	storage.Getter
	//Exists returns true if resource exists
	Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error)

	//Download download bytes
	Download(ctx context.Context, object storage.Object, options ...storage.Option) ([]byte, error)

	//DownloadWithURL download bytes for URL
	DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) ([]byte, error)

	storage.Copier
	storage.Mover

	//Initialises manager for baseURL with storage options (i.e. auth)
	Init(ctx context.Context, baseURL string, options ...storage.Option) error

	//NewWriter creates an upload writer
	NewWriter(ctx context.Context, URL string, mode os.FileMode, options ...storage.Option) (io.WriteCloser, error)

	//Closes all active managers
	CloseAll() error
	//Closes matched active manager
	Close(baseURL string) error

	//ErrorCode returns an error code or zero
	ErrorCode(scheme string, err error) int
}

//Service implementation
type service struct {
	faker    bool
	mutex    *sync.RWMutex
	managers map[string]storage.Manager
}

func (s *service) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return err
	}
	return manager.Upload(ctx, URL, mode, reader, options...)
}

func (s *service) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return err
	}
	return manager.Delete(ctx, URL, options...)
}

func (s *service) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return err
	}
	return manager.Create(ctx, URL, mode, isDir, options...)
}

func (s *service) Object(ctx context.Context, URL string, options ...storage.Option) (storage.Object, error) {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return nil, err
	}
	if getter, ok := manager.(storage.Getter); ok {
		return getter.Object(ctx, URL, options...)
	}
	return s.object(ctx, manager, URL, options...)
}

func (s *service) object(ctx context.Context, manager storage.Manager, URL string, options ...storage.Option) (storage.Object, error) {
	options = append(options, option.NewPage(0, 1))
	objects, err := manager.List(ctx, URL, options...)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("%v: not found", URL)
	}
	return objects[0], nil
}

func (s *service) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {
	URL = url.Normalize(URL, file.Scheme)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return false, err
	}
	return s.exists(ctx, manager, URL, options...)
}

func (s *service) exists(ctx context.Context, manager storage.Manager, URL string, options ...storage.Option) (bool, error) {
	if checker, ok := manager.(storage.Checker); ok {
		return checker.Exists(ctx, URL, options...)
	}
	options = append(options, option.NewPage(0, 1))
	objects, err := s.List(ctx, URL, options...)
	if err != nil {
		return false, nil
	}
	return len(objects) > 0, nil
}

func (s *service) Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return s.OpenURL(ctx, object.URL(), options...)
}

func (s *service) OpenURL(ctx context.Context, URL string, options ...storage.Option) (reader io.ReadCloser, err error) {
	URL = url.Normalize(URL, file.Scheme)
	var modifier option.Modifier
	option.Assign(options, &modifier)
	manager, err := s.manager(ctx, URL, options)
	if err != nil {
		return nil, err
	}
	reader, err = manager.OpenURL(ctx, URL, options...)
	if modifier == nil || err != nil {
		return reader, err
	}
	_, URLPath := url.Base(URL, file.Scheme)
	_, name := path.Split(URLPath)
	info := file.NewInfo(name, 0, file.DefaultFileOsMode, time.Now(), false)
	_, reader, err = modifier(info, reader)
	return reader, err
}

func (s *service) Download(ctx context.Context, object storage.Object, options ...storage.Option) ([]byte, error) {
	reader, err := s.Open(ctx, object, options...)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func (s *service) DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) ([]byte, error) {
	reader, err := s.OpenURL(ctx, URL, options...)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func (s *service) newManager(ctx context.Context, scheme string, options ...storage.Option) (storage.Manager, error) {
	if s.faker {
		scheme = mem.Scheme
	}
	provider, err := GetRegistry().Get(scheme)
	if err != nil {
		return nil, err
	}
	return provider(options...)
}

//Init initialises service
func (s *service) Init(ctx context.Context, baseURL string, options ...storage.Option) error {
	baseURL = url.Normalize(baseURL, file.Scheme)
	_, err := s.manager(ctx, baseURL, options)
	return err
}

//Close closes storage manager for supplied baseURL
func (s *service) Close(baseURL string) error {
	baseURL, _ = url.Base(baseURL, file.Scheme)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	manager, ok := s.managers[baseURL]
	if !ok {
		return nil
	}
	return manager.Close()
}

//CloseAll closes all active managers
func (s *service) CloseAll() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var err error
	for _, manager := range s.managers {
		if e := manager.Close(); e != nil {
			err = e
		}
	}
	return err
}

func (s *service) IsAuthChanged(ctx context.Context, manager storage.Manager, URL string, options []storage.Option) bool {
	authTracker, ok := manager.(storage.AuthTracker)
	if !ok {
		return false
	}
	return authTracker.IsAuthChanged(ctx, URL, options)
}

func (s *service) manager(ctx context.Context, URL string, options []storage.Option) (storage.Manager, error) {
	scheme := url.Scheme(URL, file.Scheme)
	noCache := &option.NoCache{}
	options, _ = option.Assign(options, &noCache)
	if noCache.Source == option.NoCacheBaseURL {
		return s.newManager(ctx, scheme, options...)
	}

	key, _ := url.Base(URL, scheme)
	extURL := url.SchemeExtensionURL(URL)
	key += extURL

	if extURL != "" {
		if extScheme := url.Scheme(extURL, file.Scheme); extScheme != scheme {
			extManager, err := s.manager(ctx, extURL, options)
			if err != nil {
				return nil, err
			}
			options = append(options, extManager)
		}
	}
	s.mutex.RLock()
	result, ok := s.managers[key]
	s.mutex.RUnlock()
	if ok {
		if !s.IsAuthChanged(ctx, result, URL, options) {
			return result, nil
		}
		_ = result.Close()

	}
	manager, err := s.newManager(ctx, scheme, options...)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err == nil {
		s.managers[key] = manager
	}
	return manager, err
}

//ErrorCode return error code
func (s *service) ErrorCode(scheme string, err error) int {
	if err == nil {
		return 0
	}
	manager, e := s.newManager(context.Background(), scheme)
	if e != nil {
		return 0
	}
	if coder, ok := manager.(storage.ErrorCoder); ok {
		return coder.ErrorCode(err)
	}
	return 0
}

func newService(faker bool) *service {
	return &service{
		faker:    faker,
		mutex:    &sync.RWMutex{},
		managers: make(map[string]storage.Manager),
	}
}

//New returns a abstract storage service
func New() Service {
	return newService(false)
}
