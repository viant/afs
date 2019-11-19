package storage

import (
	"context"
	"io"
	"os"
)

//Manager represents storage manager
type Manager interface {
	Lister
	Downloader
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

//Downloader represents a downloader
type Downloader interface {
	//Download returns reader for downloaded storage object
	Download(ctx context.Context, object Object, options ...Option) (io.ReadCloser, error)

	//Download returns reader for downloaded storage object
	DownloadWithURL(ctx context.Context, URL string, options ...Option) (io.ReadCloser, error)
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
