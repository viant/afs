package base

import (
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

//Manager represents storager base manager
type Manager struct {
	storage.Manager
	options   []storage.Option
	scheme    string
	mutex     *sync.RWMutex
	storagers map[string]storage.Storager
	provider  func(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error)
}

//List lists content for supplied URL
func (m *Manager) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	baseURL, URLPath := url.Base(URL, m.scheme)

	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return nil, err
	}
	files, err := storager.List(ctx, URLPath, options)
	if err != nil {
		return nil, err
	}
	var objects = make([]storage.Object, len(files))
	if len(objects) == 0 {
		return objects, nil
	}
	_, isDirect := files[0].(storage.Object)

	if isDirect {
		for i := range files {
			objects[i] = files[i].(storage.Object)
		}
		return objects, nil
	}

	objects[0] = object.New(url.Join(baseURL, URLPath), files[0], nil)
	for i := 1; i < len(files); i++ {
		fileURL := url.Join(baseURL, path.Join(URLPath, files[i].Name()))
		objects[i] = object.New(fileURL, files[i], nil)
	}
	return objects, nil
}

func (m *Manager) ensureParentExists(ctx context.Context, URL string) error {
	baseURL, URLPath := url.Base(URL, m.scheme)
	parent, _ := path.Split(URLPath)
	parentURL := url.Join(baseURL, parent)
	return m.Create(ctx, parentURL, file.DefaultDirOsMode, true)
}

//Upload uploads content
func (m *Manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	baseURL, URLPath := url.Base(URL, m.scheme)
	err := m.ensureParentExists(ctx, URL)
	if err != nil {
		return err
	}
	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return storager.Upload(ctx, URLPath, mode, data, options...)
}

//Download downloads content
func (m *Manager) Download(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return m.DownloadWithURL(ctx, object.URL(), options...)
}

//DownloadWithURL downloads content
func (m *Manager) DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	baseURL, URLPath := url.Base(URL, m.scheme)

	var modifier option.Modifier
	option.Assign(options, &modifier)
	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return nil, err
	}
	reader, err := storager.Download(ctx, URLPath, options...)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

//Delete deletes locations
func (m *Manager) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	baseURL, URLPath := url.Base(URL, m.scheme)
	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return err
	}
	return storager.Delete(ctx, URLPath)
}

//Create creates a resource
func (m *Manager) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	var reader io.Reader
	options, _ = option.Assign(options, &reader)
	baseURL, URLPath := url.Base(URL, m.scheme)
	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return err
	}
	data := []byte{}
	if reader != nil {
		if data, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	}
	return storager.Create(ctx, URLPath, mode, data, isDir)
}

//Exists checks if resource exsits
func (m *Manager) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {
	baseURL, URLPath := url.Base(URL, m.scheme)
	storager, err := m.Storager(ctx, baseURL, options...)
	if err != nil {
		return false, err
	}
	return storager.Exists(ctx, URLPath)
}

//Options returns base and supplied options
func (m *Manager) Options(options []storage.Option) []storage.Option {
	result := make([]storage.Option, 0)
	result = append(result, m.options...)
	result = append(result, options...)
	return result
}

//Storager returns storager
func (m *Manager) Storager(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	m.mutex.RLock()
	baseURL, _ = url.Base(baseURL, m.scheme)
	storager, ok := m.storagers[baseURL]
	m.mutex.RUnlock()
	if ok {
		return storager, nil
	}
	options = m.Options(options)
	storager, err := m.provider(ctx, baseURL, options...)
	if err != nil {
		return nil, err
	}
	m.storagers[baseURL] = storager
	return storager, nil
}

//Close closes storagers
func (m *Manager) Close() error {
	var err error
	for _, storager := range m.storagers {
		if e := storager.Close(); e != nil {
			err = e
		}
	}
	return err
}

//Scheme returns scheme
func (m *Manager) Scheme() string {
	return m.scheme
}

//New creates base storager base Manager
func New(manager storage.Manager, scheme string, provider func(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error), options []storage.Option) *Manager {
	return &Manager{
		Manager:   manager,
		scheme:    scheme,
		mutex:     &sync.RWMutex{},
		storagers: make(map[string]storage.Storager),
		provider:  provider,
		options:   options,
	}
}
