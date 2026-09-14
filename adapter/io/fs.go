package io

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/viant/afs"
	"github.com/viant/afs/url"
)

// FS adapts a Service rooted at Root URI into an io/fs filesystem.
// Root should be a fully-qualified URI such as s3://bucket/path or gs://... .
type FS struct {
	Ctx  context.Context
	Svc  afs.Service
	Root string
}

// NewFS constructs a new adapter filesystem.
func NewFS(ctx context.Context, svc afs.Service, root string) *FS {
	return &FS{Ctx: ctx, Svc: svc, Root: strings.TrimRight(root, "/")}
}

// Open implements fs.FS. It supports opening both files and directories.
func (a *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	uri := a.Root
	if name != "." {
		uri = a.join(name)
	}
	object, err := a.Svc.Object(a.Ctx, uri)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if object.IsDir() {
		return a.openDir(object)
	}
	data, err := a.Svc.Download(a.Ctx, object)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	return &file{
		name: path.Base(uri),
		data: data,
		info: object,
	}, nil
}

func (a *FS) openDir(object storage.Object) (fs.File, error) {
	entries, err := a.Svc.List(a.Ctx, object.URL())
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	dirents := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		if url.Equals(object.URL(), e.URL()) {
			continue
		}
		dirents = append(dirents, &dirEntry{i: e})
	}
	return &dir{
		name: object.Name(),
		ents: dirents,
		info: object,
	}, nil
}

func (a *FS) join(name string) string {
	return a.Root + "/" + strings.TrimLeft(name, "/")
}

// file is an in-memory fs.File backed by a byte slice.
type file struct {
	name string
	data []byte
	off  int
	info fs.FileInfo
}

func (f *file) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *file) Read(p []byte) (int, error) {
	if f.off >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	return n, nil
}
func (f *file) Close() error { return nil }

// dir is a directory implementing fs.ReadDirFile.
type dir struct {
	name string
	ents []fs.DirEntry
	pos  int
	info fs.FileInfo
}

func (d *dir) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *dir) Read(_ []byte) (int, error) { return 0, io.EOF }
func (d *dir) Close() error               { return nil }
func (d *dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.ents) {
		if n <= 0 {
			return []fs.DirEntry{}, nil
		}
		return nil, io.EOF
	}
	if n <= 0 || d.pos+n > len(d.ents) {
		n = len(d.ents) - d.pos
	}
	out := d.ents[d.pos : d.pos+n]
	d.pos += n
	return out, nil
}

// dirEntry implements fs.DirEntry using info.
type dirEntry struct{ i fs.FileInfo }

func (e *dirEntry) Name() string { return e.i.Name() }
func (e *dirEntry) IsDir() bool  { return e.i.IsDir() }
func (e *dirEntry) Type() fs.FileMode {
	if e.i.IsDir() {
		return fs.ModeDir
	}
	return 0
}
func (e *dirEntry) Info() (fs.FileInfo, error) { return e.i, nil }
