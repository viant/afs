package storage

import (
	"context"
	"io"
	"os"
)

//Storager represents path oriented storage service
type Storager interface {
	io.Closer

	//Exists returns true if location exists
	Exists(ctx context.Context, location string, options ...Option) (bool, error)

	//List lists location assets
	List(ctx context.Context, location string, options ...Option) ([]os.FileInfo, error)

	//Get returns a file info for supplied location
	Get(ctx context.Context, location string, options ...Option) (os.FileInfo, error)

	//Open returns a reader closer for supplied resources
	Open(ctx context.Context, location string, options ...Option) (io.ReadCloser, error)

	//Upload uploads
	Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...Option) error

	//Create create file or directory
	Create(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, isDir bool, options ...Option) error

	//Delete deletes locations
	Delete(ctx context.Context, location string, options ...Option) error
}
