package mem

import (
	"bytes"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

//File represents in memory file
type File struct {
	storage.Object
	content       []byte
	modTime       time.Time
	downloadError error
	uploadError   error
	readerError   error
	generation    int64
}

//NewReader return new Reader
func (f *File) NewReader() io.ReadCloser {
	var reader io.Reader = bytes.NewReader(f.content)
	if f.readerError != nil {
		reader = &fakeReader{error: f.readerError}
	}
	return ioutil.NopCloser(reader)
}

//SetErrors sets test errors
func (f *File) SetErrors(errors ...*option.Error) {
	if len(errors) > 0 {
		for i := range errors {
			switch strings.ToLower(errors[i].Type) {
			case option.ErrorTypeDownload:
				f.downloadError = errors[i].Error
			case option.ErrorTypeUpload:
				f.uploadError = errors[i].Error
			case option.ErrorTypeReader:
				f.readerError = errors[i].Error
			}
		}
	}
}

//NewFile create a file
func NewFile(URL string, mode os.FileMode, content []byte, modTime time.Time) *File {
	baseURL, URLPath := Split(URL)
	URL = url.Join(baseURL, URLPath)
	_, name := path.Split(URLPath)
	info := file.NewInfo(name, 0, mode, modTime, false)
	result := &File{
		content: content,
	}
	result.Object = object.New(URL, info, result)
	return result
}
