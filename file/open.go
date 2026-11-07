package file

import (
	"context"
	"io"
	"os"

	"github.com/viant/afs/storage"
)

// Open opens the supplied object for reading.
func Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return OpenURL(ctx, object.URL(), options...)
}

// OpenURL opens the supplied URL for reading.
func OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	return os.Open(Path(URL))
}
