package io

import (
	"context"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

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
	clean := path.Clean(name)
	if clean == "." || clean == "/" {
		return a.openDir(a.Root, "")
	}
	uri := a.join(clean)

	// Try directory first; if it fails, try file.
	if d, err := a.openDir(uri, url.Name(uri)); err == nil {
		return d, nil
	}
	data, err := a.Svc.DownloadWithURL(a.Ctx, uri)
	if err != nil {
		return nil, err
	}
	return &file{
		name: path.Base(uri),
		data: data,
		info: &info{name: path.Base(uri), size: int64(len(data)), mod: time.Time{}, dir: false},
	}, nil
}

func (a *FS) openDir(uri, name string) (fs.File, error) {
	entries, err := a.Svc.List(a.Ctx, uri)
	if err != nil {
		return nil, err
	}
	dirents := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		dirents = append(dirents, &dirEntry{i: &info{name: e.Name(), size: e.Size(), mod: e.ModTime(), dir: e.IsDir()}})
	}
	return &dir{
		name: name,
		ents: dirents,
		info: &info{name: name, dir: true},
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
	info *info
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
	info *info
}

func (d *dir) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *dir) Read(_ []byte) (int, error) { return 0, io.EOF }
func (d *dir) Close() error               { return nil }
func (d *dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.ents) {
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
type dirEntry struct{ i *info }

func (e *dirEntry) Name() string { return e.i.name }
func (e *dirEntry) IsDir() bool  { return e.i.dir }
func (e *dirEntry) Type() fs.FileMode {
	if e.i.dir {
		return fs.ModeDir
	}
	return 0
}
func (e *dirEntry) Info() (fs.FileInfo, error) { return e.i, nil }

// info implements fs.FileInfo.
type info struct {
	name string
	size int64
	mod  time.Time
	dir  bool
}

func (i *info) Name() string { return i.name }
func (i *info) Size() int64  { return i.size }
func (i *info) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (i *info) ModTime() time.Time { return i.mod }
func (i *info) IsDir() bool        { return i.dir }
func (i *info) Sys() interface{}   { return nil }
