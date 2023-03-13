package cache

import (
	"github.com/viant/afs"
	"github.com/viant/afs/storage"
)

var singleton afs.Service

//Singleton returns caching Service for specified URL
func Singleton(URL string, opts ...storage.Option) afs.Service {
	if singleton != nil {
		return singleton
	}
	singleton = New(URL, afs.New(), opts...)
	return singleton
}
