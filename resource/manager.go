package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/viant/afs"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

var (
	_ storage.Manager        = (*manager)(nil)
	_ storage.Getter         = (*manager)(nil)
	_ storage.Checker        = (*manager)(nil)
	_ storage.WriterProvider = (*manager)(nil)
)

type manager struct {
	scheme  string
	roots   map[string]root
	service afs.Service
}

func (m *manager) Scheme() string { return m.scheme }

// Close does not close caller-owned filesystems or the backing AFS service.
func (m *manager) Close() error { return nil }

// mappedObject preserves native metadata and Unwrap while retaining a virtual URL.
type mappedObject struct {
	storage.Object
	virtual string
}

func (o *mappedObject) URL() string { return o.virtual }

func (m *manager) Object(ctx context.Context, URL string, options ...storage.Option) (storage.Object, error) {
	l, err := m.resolve(ctx, URL)
	if err != nil {
		return nil, err
	}
	var result storage.Object
	if l.root.FS != nil {
		var info fs.FileInfo
		info, err = fs.Stat(l.root.FS, l.target)
		if err == nil {
			result = object.New(URL, info, nil)
		}
	} else {
		result, err = m.service.Object(ctx, l.target, options...)
	}
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if result.Mode()&fs.ModeSymlink != 0 {
		return nil, ErrBoundary
	}
	return &mappedObject{Object: result, virtual: URL}, nil
}

func (m *manager) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {
	l, err := m.resolve(ctx, URL)
	if err != nil {
		return false, err
	}
	if l.root.FS == nil {
		exists, err := m.service.Exists(ctx, l.target, options...)
		if err == nil {
			err = ctx.Err()
		}
		return exists, err
	}
	_, err = m.Object(ctx, URL, options...)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (m *manager) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	l, err := m.resolve(ctx, URL)
	if err != nil {
		return nil, err
	}
	if l.root.FS == nil {
		entries, err := m.service.List(ctx, l.target, options...)
		if err != nil {
			return nil, err
		}
		result := make([]storage.Object, 0, len(entries))
		for _, entry := range entries {
			virtual, err := l.virtualObjectURL(entry.URL())
			if err != nil {
				return nil, err
			}
			if entry.URL() == l.target {
				virtual = URL
			}
			if _, err = m.resolve(ctx, virtual); err != nil {
				return nil, err
			}
			if entry.Mode()&fs.ModeSymlink != 0 {
				return nil, ErrBoundary
			}
			result = append(result, &mappedObject{Object: entry, virtual: virtual})
		}
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}
	parent, err := m.Object(ctx, URL, options...)
	if err != nil {
		return nil, err
	}
	if !parent.IsDir() {
		return []storage.Object{parent}, nil
	}
	entries, err := fs.ReadDir(l.root.FS, l.target)
	if err != nil {
		return nil, err
	}
	// Like file/embed, directory listings include the directory itself.
	result := []storage.Object{parent}
	match, page := option.GetListOptions(options)
	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil, ErrBoundary
		}
		if !match(l.name, info) {
			continue
		}
		page.Increment()
		if page.ShallSkip() {
			continue
		}
		virtual := l.childURL(entry.Name())
		if _, err = m.resolve(ctx, virtual); err != nil {
			return nil, err
		}
		result = append(result, object.New(virtual, info, nil))
		if page.HasReachedLimit() {
			break
		}
	}
	return result, ctx.Err()
}

func (m *manager) Open(ctx context.Context, obj storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return m.OpenURL(ctx, obj.URL(), options...)
}

func (m *manager) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	l, err := m.resolve(ctx, URL)
	if err != nil {
		return nil, err
	}
	var reader io.ReadCloser
	if l.root.FS != nil {
		reader, err = l.root.FS.Open(l.target)
	} else {
		var modifier option.Modifier
		forward, _ := option.Assign(options, &modifier)
		reader, err = m.service.OpenURL(ctx, l.target, forward...)
	}
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		_ = reader.Close()
		return nil, err
	}
	return &contextReadCloser{Reader: contextReader{ctx: ctx, reader: reader}, Closer: reader}, nil
}

func (m *manager) writable(ctx context.Context, URL string, options []storage.Option) (location, error) {
	l, err := m.resolve(ctx, URL)
	if err != nil {
		return l, err
	}
	if l.root.FS != nil {
		return l, fmt.Errorf("%w: %s", ErrReadOnly, URL)
	}
	// Native link options can carry a second pathname outside the mapped root.
	for _, value := range options {
		switch value.(type) {
		case object.Link, *object.Link:
			return l, fmt.Errorf("%w: link creation is not supported", ErrBoundary)
		}
	}
	return l, nil
}

func (m *manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	l, err := m.writable(ctx, URL, options)
	if err != nil {
		return err
	}
	if reader != nil {
		reader = contextReader{ctx: ctx, reader: reader}
	}
	if err = m.service.Upload(ctx, l.target, mode, reader, options...); err != nil {
		return err
	}
	return ctx.Err()
}

func (m *manager) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	l, err := m.writable(ctx, URL, options)
	if err != nil {
		return err
	}
	if err = m.service.Create(ctx, l.target, mode, isDir, options...); err != nil {
		return err
	}
	return ctx.Err()
}

func (m *manager) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	l, err := m.writable(ctx, URL, options)
	if err != nil {
		return err
	}
	if err = m.service.Delete(ctx, l.target, options...); err != nil {
		return err
	}
	return ctx.Err()
}

func (m *manager) NewWriter(ctx context.Context, URL string, mode os.FileMode, options ...storage.Option) (io.WriteCloser, error) {
	l, err := m.writable(ctx, URL, options)
	if err != nil {
		return nil, err
	}
	writer, err := m.service.NewWriter(ctx, l.target, mode, options...)
	if err != nil {
		return nil, err
	}
	return &contextWriter{ctx: ctx, WriteCloser: writer}, nil
}
