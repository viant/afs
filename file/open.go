package file

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"os"
	"path"
)

//Open downloads TestContent for the supplied object
func Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return OpenURL(ctx, object.URL(), options...)
}

//OpenURL downloads content for the supplied object
func OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	filePath := Path(URL)
	parent, _ := path.Split(filePath)
	if err := EnsureParentPathExists(parent, DefaultDirOsMode); err != nil {
		return nil, err
	}
	return os.Open(filePath)
}
