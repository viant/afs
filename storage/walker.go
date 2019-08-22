package storage

import (
	"context"
	"io"
	"os"
)

//OnVisit represents on location visit handler
type OnVisit func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error)

//Walker represents abstract storage walker
type Walker interface {
	//Walk traverses URL and calls handler on all file or folder
	Walk(ctx context.Context, URL string, handler OnVisit, options ...Option) error
}
