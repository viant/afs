package storage

import (
	"context"
	"io"
	"os"
)

//Upload uploads content
type Upload func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error

//Uploader represents an uploader
type Uploader interface {
	//Upload uploads provided reader content for supplied storage object.
	Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...Option) error
}

//BatchUploader represents a batch uploader
type BatchUploader interface {
	//Uploader returns upload handler, and upload closer for batch upload or error
	Uploader(ctx context.Context, URL string, options ...Option) (Upload, io.Closer, error)
}
