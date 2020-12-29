package storage

import (
	"context"
	"io"
	"os"
)

//Manager represents storage manager
type Manager interface {
	Lister
	Opener
	Uploader
	Deleter
	Creator
	io.Closer
	Scheme() string
}

//Lister represents asset lister
type Lister interface {
	//List returns a list of object for supplied url
	List(ctx context.Context, URL string, options ...Option) ([]Object, error)
}

//Getter represents asset getter
type Getter interface {
	//List returns a list of object for supplied url
	Object(ctx context.Context, URL string, options ...Option) (Object, error)
}

//Opener represents a downloader
type Opener interface {
	//Open returns reader for downloaded storage object
	Open(ctx context.Context, object Object, options ...Option) (io.ReadCloser, error)

	//Open returns reader for downloaded storage object
	OpenURL(ctx context.Context, URL string, options ...Option) (io.ReadCloser, error)
}

//Deleter represents a deleter
type Deleter interface {
	//Delete removes passed in storage object
	Delete(ctx context.Context, URL string, options ...Option) error
}

//Creator represents a creator
type Creator interface {
	//CreateBucket creates a bucket
	Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...Option) error
}

//ErrorCoder represents error coder
type ErrorCoder interface {
	//ErrorCode returns an error code
	ErrorCode(err error) int
}
