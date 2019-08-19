package storage

import "os"

//Object represents a storage object
type Object interface {
	os.FileInfo
	//URL return storage url
	URL() string

	//Wrap wraps source storage object
	Wrap(source interface{})
	//Unwrap unwraps source storage object into provided target.
	Unwrap(target interface{}) error
	//FileInfo return file info
}
