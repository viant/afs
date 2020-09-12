package storage

import (
	"context"
	"io"
	"os"
)

//WriterProvider represents writer provider
type WriterProvider interface {
	NewWriter(ctx context.Context, URL string, mode os.FileMode, options ...Option) (io.WriteCloser, error)
}
