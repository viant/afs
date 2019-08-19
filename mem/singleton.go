package mem

import "github.com/viant/afs/storage"

var singleton *manager

//Singleton returns singleton manager
func Singleton(options ...storage.Option) storage.Manager {
	if singleton != nil {
		return singleton
	}
	singleton = newManager(options...)
	return singleton
}

//ResetSingleton rest singleton
func ResetSingleton(options ...storage.Option) {
	singleton = newManager(options...)
}
