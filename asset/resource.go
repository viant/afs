package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
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
	Dir      bool
	Link     string
	Name     string
	Mode     os.FileMode
	ModTime  *time.Time
	Data     []byte
	FileInfo os.FileInfo
}

//Info returns file info
func (r Resource) Info() os.FileInfo {
	if r.FileInfo != nil {
		return r.FileInfo
	}
	modTime := time.Now()
	if r.ModTime != nil {
		modTime = *r.ModTime
	}
	name := r.Name
	if strings.Contains(name, "/") {
		_, name = path.Split(r.Name)
	}
	r.FileInfo = file.NewInfo(name, int64(len(r.Data)), r.Mode, modTime, r.Dir, object.NewLink(r.Link, r.Link, nil))
	return r.FileInfo
}

//Reader returns a reader
func (r Resource) Reader() io.Reader {
	if r.Dir {
		return nil
	}
	return bytes.NewReader(r.Data)
}

//MergeFrom merges into supplied resource
func (r *Resource) MergeFrom(resource *Resource) error {
	if r.Dir != resource.Dir {
		not := ""
		if !resource.Dir {
			not = "not "
		}
		return fmt.Errorf("%v: is %vdirectory", resource.Name, not)
	}
	r.Data = resource.Data
	r.Mode = resource.Mode
	if resource.FileInfo != nil {
		r.FileInfo = resource.FileInfo
	}
	return nil
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

//NewGzFile create a gz file resource
func NewGzFile(name string, data []byte, mode os.FileMode) *Resource {
	buffer := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(buffer)
	io.Copy(gzWriter, bytes.NewReader(data))
	gzWriter.Flush()
	gzWriter.Close()
	return New(name, mode, false, "", buffer.Bytes())
}

//New creates an asset
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
