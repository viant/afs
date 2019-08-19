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
	Exists(ctx context.Context, location string) (bool, error)

	//List lists location assets
	List(ctx context.Context, location string, options ...Option) ([]os.FileInfo, error)

	//Download feches content for supplied location
	Download(ctx context.Context, location string, options ...Option) (io.ReadCloser, error)

	//Upload uploads
	Upload(ctx context.Context, destination string, mode os.FileMode, content []byte, options ...Option) error

	//Create create file or directory
	Create(ctx context.Context, destination string, mode os.FileMode, content []byte, isDir bool, options ...Option) error

	//Delete deletes locations
	Delete(ctx context.Context, location string) error
}
