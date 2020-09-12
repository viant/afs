package file

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"os"
)

//Open downloads TestContent for the supplied object
func Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return OpenURL(ctx, object.URL(), options...)
}

//OpenURL downloads content for the supplied object
func OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	filePath := Path(URL)
	return os.Open(filePath)
}
