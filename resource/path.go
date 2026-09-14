package resource

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type root struct {
	Root
	local string
}
type location struct {
	root                  root
	name, target, virtual string
}

func parseURL(value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed URL", ErrBoundary)
	}
	if u.Scheme == "" || !strings.HasPrefix(value, u.Scheme+"://") || u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(value, "#") || strings.ContainsAny(u.Host, `\%`) {
		return nil, fmt.Errorf("%w: invalid URL", ErrBoundary)
	}
	return u, nil
}

// Decode exactly once, rejecting encoded separators and further escapes rather
// than permitting a downstream provider to interpret a different path.
func relativePath(u *url.URL) (string, error) {
	escaped := strings.ToLower(u.EscapedPath())
	if strings.Contains(escaped, "%2f") || strings.Contains(escaped, "%5c") {
		return "", fmt.Errorf("%w: encoded separator", ErrBoundary)
	}
	value := strings.TrimPrefix(u.Path, "/")
	value = strings.TrimSuffix(value, "/")
	if value == "" {
		return ".", nil
	}
	if !fs.ValidPath(value) || strings.ContainsAny(value, "\\%:\x00") {
		return "", fmt.Errorf("%w: invalid relative path", ErrBoundary)
	}
	return value, nil
}

func (m *manager) resolve(ctx context.Context, value string) (location, error) {
	if err := ctx.Err(); err != nil {
		return location{}, err
	}
	u, err := parseURL(value)
	if err != nil {
		return location{}, err
	}
	if u.Scheme != m.scheme {
		return location{}, fmt.Errorf("%w: scheme %q", ErrUnknownNamespace, u.Scheme)
	}
	r, ok := m.roots[u.Host]
	if !ok {
		return location{}, fmt.Errorf("%w: %s://%s", ErrUnknownNamespace, m.scheme, u.Host)
	}
	name, err := relativePath(u)
	if err != nil {
		return location{}, err
	}
	l := location{root: r, name: name, virtual: value}
	if r.FS != nil {
		l.target = path.Join(r.Path, name)
	} else {
		l.target = r.URL
		if name != "." {
			l.target += "/" + name
		}
		if err = r.checkLocal(name); err != nil {
			return location{}, err
		}
	}
	return l, nil
}

// Native AFS file operations are pathname based. Reject every existing symlink
// component, including ancestors; absent tails are allowed for create/upload.
// This is not a sandbox against concurrent external filesystem replacement.
func (r root) checkLocal(name string) error {
	if r.local == "" {
		return nil
	}
	target := r.local
	if name != "." {
		target = filepath.Join(target, filepath.FromSlash(name))
	}
	for current := target; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symlink component", ErrBoundary)
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}

func (l location) childURL(name string) string {
	return strings.TrimRight(l.virtual, "/") + "/" + url.PathEscape(name)
}

func (l location) virtualObjectURL(target string) (string, error) {
	base := strings.TrimRight(l.root.URL, "/")
	if target == base || strings.TrimRight(target, "/") == base {
		return l.namespaceURL(), nil
	}
	if !strings.HasPrefix(target, base+"/") {
		return "", fmt.Errorf("%w: backend returned URL outside root", ErrBoundary)
	}
	relative := strings.TrimPrefix(target, base+"/")
	if !fs.ValidPath(relative) || strings.ContainsAny(relative, "\\%:\x00") {
		return "", ErrBoundary
	}
	parts := strings.Split(relative, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return l.namespaceURL() + "/" + strings.Join(parts, "/"), nil
}

func (l location) namespaceURL() string {
	u, _ := url.Parse(l.virtual)
	return u.Scheme + "://" + u.Host
}
