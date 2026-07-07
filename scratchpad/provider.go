package scratchpad

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	neturl "net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/viant/afs"
	"github.com/viant/afs/base"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	afsurl "github.com/viant/afs/url"
)

var (
	registryMu       sync.RWMutex
	registeredConfig Config
)

func Register(opts ...Option) {
	cfg := Config{RootURITemplate: DefaultRootURITemplate}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	registryMu.Lock()
	registeredConfig = cfg
	registryMu.Unlock()
	afs.GetRegistry().Register(Scheme, Provider)
}

func Provider(options ...storage.Option) (storage.Manager, error) {
	registryMu.RLock()
	cfg := registeredConfig
	registryMu.RUnlock()
	option.Assign(options, &cfg)
	return newManager(New(WithConfig(cfg))), nil
}

type manager struct {
	*base.Manager
	service *Service
}

func newManager(service *Service) *manager {
	result := &manager{service: service}
	baseMgr := base.New(result, Scheme, result.provider, nil)
	result.Manager = baseMgr
	return result
}

func (m *manager) provider(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	cfg := Config{}
	option.Assign(options, &cfg)
	service := m.service
	if cfg.RootURITemplate != "" || cfg.UserID != "" || cfg.FS != nil || len(cfg.AllowedTargetSchemes) > 0 {
		service = New(WithConfig(cfg))
	}
	return &storager{service: service, baseURL: baseURL}, nil
}

type storager struct {
	base.Storager
	service *Service
	baseURL string
}

func (s *storager) Close() error { return nil }

func (s *storager) Exists(ctx context.Context, location string, options ...storage.Option) (bool, error) {
	_, err := s.Get(ctx, location, options...)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	output, err := s.service.List(ctx)
	if err != nil {
		return nil, err
	}
	kind := s.kind()
	var result []os.FileInfo
	for _, entry := range output.Entries {
		if kind == "artifact" && !strings.HasPrefix(entry.Key, "artifact/") {
			continue
		}
		name := entry.Key
		if kind == "artifact" {
			name = strings.TrimPrefix(entry.Key, "artifact/")
		}
		if location != "" && !strings.HasPrefix(name, strings.Trim(location, "/")) {
			continue
		}
		modTime := time.Now()
		if entry.UpdatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, entry.UpdatedAt); err == nil {
				modTime = parsed
			}
		}
		result = append(result, file.NewInfo(name, 0, file.DefaultFileOsMode, modTime, false))
	}
	return result, nil
}

func (s *storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	key, err := s.key(location)
	if err != nil {
		return nil, err
	}
	if s.kind() == "artifact" {
		artifact, err := s.service.FetchArtifact(ctx, key)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSpace(artifact.Name)
		if name == "" {
			name = key
		}
		return file.NewInfo(name, 0, file.DefaultFileOsMode, time.Now(), false), nil
	}
	note, err := s.service.Fetch(ctx, key)
	if err != nil {
		return nil, err
	}
	modTime := time.Now()
	if note.UpdatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, note.UpdatedAt); err == nil {
			modTime = parsed
		}
	}
	return file.NewInfo(path.Base(key), int64(len(note.Body)), file.DefaultFileOsMode, modTime, false), nil
}

func (s *storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	key, err := s.key(location)
	if err != nil {
		return nil, err
	}
	if s.kind() == "artifact" {
		_, reader, err := s.service.OpenArtifact(ctx, key)
		return reader, err
	}
	note, err := s.service.Fetch(ctx, key)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(note.Body)), nil
}

func (s *storager) Upload(ctx context.Context, location string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	key, err := s.key(location)
	if err != nil {
		return err
	}
	var data []byte
	if reader != nil {
		if data, err = io.ReadAll(reader); err != nil {
			return err
		}
	}
	description := key
	if s.kind() == "artifact" {
		key = ArtifactKey(key)
		description = "Artifact " + strings.TrimPrefix(key, "artifact/")
	}
	_, err = s.service.Memorize(ctx, &MemorizeInput{
		Key:         key,
		Description: description,
		Body:        string(data),
	})
	return err
}

func (s *storager) Create(ctx context.Context, location string, mode os.FileMode, reader io.Reader, isDir bool, options ...storage.Option) error {
	if isDir {
		return nil
	}
	return s.Upload(ctx, location, mode, reader, options...)
}

func (s *storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	key, err := s.key(location)
	if err != nil {
		return err
	}
	if s.kind() == "artifact" {
		key = ArtifactKey(key)
	}
	return s.service.Delete(ctx, key)
}

func (s *storager) key(location string) (string, error) {
	location = strings.Trim(location, "/")
	if location == "" {
		return "", fmt.Errorf("scratchpad key is required")
	}
	value, err := neturl.PathUnescape(location)
	if err != nil {
		return "", fmt.Errorf("invalid scratchpad key: %w", err)
	}
	return strings.Trim(value, "/"), nil
}

func (s *storager) kind() string {
	u, err := neturl.Parse(s.baseURL)
	if err == nil {
		if u.Host != "" {
			return strings.ToLower(u.Host)
		}
		if trimmed := strings.Trim(u.Path, "/"); trimmed != "" {
			parts := strings.Split(trimmed, "/")
			return strings.ToLower(parts[0])
		}
	}
	return "note"
}

func ArtifactSourceURL(artifact Artifact) string {
	data, _ := json.Marshal(artifact)
	return string(bytes.TrimSpace(data))
}

func ObjectForURL(url string) storage.Object {
	_, urlPath := afsurl.Base(url, Scheme)
	return object.New(url, file.NewInfo(path.Base(urlPath), 0, file.DefaultFileOsMode, time.Now(), false), nil)
}
