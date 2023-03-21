package storage

import (
	"os"
	"sync"
)

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

//Objects represents synchromized object collection wrapper
type Objects struct {
	ptr *[]Object
	mux sync.Mutex
}

//Append appens object
func (s *Objects) Append(object Object) {
	s.mux.Lock()
	*s.ptr = append(*s.ptr, object)
	s.mux.Unlock()
}

//Objects returns objects
func (s *Objects) Objects() []Object {
	return *s.ptr
}

//NewObjects creates sycnronized objects
func NewObjects(ptr *[]Object) *Objects {
	if ptr == nil {
		obj := make([]Object, 0)
		ptr = &obj
	}
	return &Objects{ptr: ptr}
}
