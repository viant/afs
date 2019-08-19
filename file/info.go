package file

import (
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"os"
	"time"
)

type Info struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	*object.Link
}

func (i *Info) Name() string {
	return i.name
}

func (i *Info) Size() int64 {
	return i.size
}
func (i *Info) Mode() os.FileMode {
	return i.mode
}
func (i *Info) ModTime() time.Time {
	return i.modTime
}

func (i *Info) IsDir() bool {
	return i.isDir
}

func (i *Info) Sys() interface{} {
	return i.Source
}

//NewInfo returns a ew file Info
func NewInfo(name string, size int64, mode os.FileMode, modificationTime time.Time, isDir bool, options ...storage.Option) os.FileInfo {
	link := &object.Link{}
	_, _ = option.Assign(options, &link)
	if link.Source == nil && len(options) == 1 {
		link.Source = options[0]
	}
	return &Info{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modificationTime,
		isDir:   isDir,
		Link:    link,
	}
}
