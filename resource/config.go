// Package resource maps caller-defined URL schemes and authorities to exact
// filesystem roots. It supplies files, not application resource metadata.
package resource

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/viant/afs"
	"github.com/viant/afs/storage"
)

var (
	ErrUnknownNamespace = errors.New("unknown resource namespace")
	ErrBoundary         = errors.New("resource path outside boundary")
	ErrReadOnly         = errors.New("read-only resource root")
)

// Root selects exactly one of FS or URL. Path is an exact subtree within FS
// (empty means "."). URL is an absolute native AFS directory URL. FS must
// provide its own confinement (embed.FS does; os.DirFS does not).
// Use a file URL for local paths so the mapping can check symlinks.
type Root struct {
	FS   fs.FS
	Path string
	URL  string
}

// Config is copied by NewProvider. The filesystem and backing service are
// borrowed and must remain valid and safe for concurrent use. Service is required
// for URL roots; the caller retains responsibility for closing it.
type Config struct {
	Scheme  string
	Roots   map[string]Root
	Service afs.Service
}

var schemePattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]*$`)

// NewProvider builds an immutable factory suitable for afs.NewWithProviders.
// No global registration, authority discovery or network access occurs here.
// Storage options passed to the factory cannot change the configured mappings.
func NewProvider(config Config) (afs.Provider, error) {
	if !schemePattern.MatchString(config.Scheme) {
		return nil, fmt.Errorf("invalid resource scheme %q", config.Scheme)
	}
	m := &manager{scheme: config.Scheme, service: config.Service, roots: make(map[string]root, len(config.Roots))}
	for authority, value := range config.Roots {
		u, err := parseURL(config.Scheme + "://" + authority + "/")
		if err != nil || u.Host != authority || authority == "" {
			return nil, fmt.Errorf("invalid resource authority %q", authority)
		}
		r := root{Root: value}
		if (value.FS == nil) == (value.URL == "") {
			return nil, fmt.Errorf("resource %q needs exactly one of FS or URL", authority)
		}
		if value.FS != nil {
			if r.Path == "" {
				r.Path = "."
			}
			if !fs.ValidPath(r.Path) || strings.ContainsAny(r.Path, `\%:`) {
				return nil, fmt.Errorf("invalid filesystem root %q", r.Path)
			}
		} else {
			if value.Path != "" || config.Service == nil {
				return nil, fmt.Errorf("URL root %q requires Service and no Path", authority)
			}
			target, err := parseURL(value.URL)
			if err != nil || target.Scheme == config.Scheme || (target.Scheme != "file" && target.Host == "") {
				return nil, fmt.Errorf("invalid or recursive URL root %q", value.URL)
			}
			if _, err = relativePath(target); err != nil {
				return nil, err
			}
			r.URL = strings.TrimRight(value.URL, "/")
			if target.Scheme == "file" {
				if target.Host != "" && target.Host != "localhost" {
					return nil, fmt.Errorf("file root requires local authority")
				}
				if !filepath.IsAbs(target.Path) {
					return nil, fmt.Errorf("file root must be absolute")
				}
				// Resolve the caller's root once, then reject links below this path.
				r.local, err = filepath.EvalSymlinks(target.Path)
				if err != nil {
					return nil, err
				}
				r.local, err = filepath.Abs(r.local)
				if err != nil {
					return nil, err
				}
				info, statErr := os.Stat(r.local)
				if statErr != nil {
					return nil, statErr
				}
				if !info.IsDir() {
					return nil, fmt.Errorf("file root must be a directory")
				}
				r.URL = "file://localhost" + filepath.ToSlash(r.local)
			}
		}
		m.roots[authority] = r
	}
	return func(...storage.Option) (storage.Manager, error) { return m, nil }, nil
}
