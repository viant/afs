package asset

import (
	"bytes"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

//Resource represents a test resource
type Resource struct {
	Dir  bool
	Link string
	Name string
	Mode os.FileMode
	Data []byte
}

//Info returns file info
func (r *Resource) Info() os.FileInfo {
	name := r.Name
	if strings.Contains(name, "/") {
		_, name = path.Split(r.Name)
	}
	return file.NewInfo(name, int64(len(r.Data)), r.Mode, time.Now(), r.Dir, object.NewLink(r.Link, r.Link, nil))
}

//Reader returns a reader
func (r *Resource) Reader() io.Reader {
	if r.Dir {
		return nil
	}
	return bytes.NewReader(r.Data)
}

//NewFile create a file resource
func NewFile(name string, data []byte, mode os.FileMode) *Resource {
	return New(name, mode, false, "", data)
}

//NewDir create a folder resource
func NewDir(name string, mode os.FileMode) *Resource {
	return New(name, mode|os.ModeDir, true, "", nil)
}

//NewLink create a link resource
func NewLink(name, link string, mode os.FileMode) *Resource {
	return New(name, mode|os.ModeSymlink, false, link, nil)
}

//NewAsset creates a new asset
func New(name string, mode os.FileMode, dir bool, link string, data []byte) *Resource {
	if mode == 0 {
		mode = 0744
	}
	return &Resource{
		Name: name,
		Data: data,
		Mode: mode,
		Dir:  dir,
		Link: link,
	}
}
