package file

import (
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"os"
	"time"
)

//Info represents a file info
type Info struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	*object.Link
}

//Name returns a name
func (i *Info) Name() string {
	return i.name
}

//Size returns file size
func (i *Info) Size() int64 {
	return i.size
}

//Mode returns file mode
func (i *Info) Mode() os.FileMode {
	return i.mode
}

//ModTime returns modification time
func (i *Info) ModTime() time.Time {
	return i.modTime
}

//IsDir returns true if resoruce is directory
func (i *Info) IsDir() bool {
	return i.isDir
}

//Sys returns sys object
func (i *Info) Sys() interface{} {
	return i.Source
}

//NewInfo returns a ew file Info
func NewInfo(name string, size int64, mode os.FileMode, modificationTime time.Time, isDir bool, options ...storage.Option) os.FileInfo {
	link := &object.Link{}
	option.Assign(options, &link)
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

//AdjustInfoSize adjust file info size
func AdjustInfoSize(info os.FileInfo, size int) os.FileInfo {
	if int(info.Size()) == size {
		return info
	}
	if fileInfo, ok := info.(*Info); ok {
		fileInfo.size = int64(size)
	} else {
		info = NewInfo(info.Name(), int64(size), info.Mode().Perm(), info.ModTime(), info.IsDir(), info.Sys())
	}
	return info
}
