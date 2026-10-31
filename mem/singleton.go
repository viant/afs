package mem

import (
	"sync"

	"github.com/viant/afs/storage"
)

var singleton *manager
var singletonMu sync.RWMutex

// Singleton returns singleton manager
func Singleton(options ...storage.Option) storage.Manager {
	singletonMu.RLock()
	current := singleton
	singletonMu.RUnlock()
	if current != nil {
		return current
	}
	singletonMu.Lock()
	defer singletonMu.Unlock()
	if singleton == nil {
		singleton = newManager(options...)
	}
	return singleton
}

// ResetSingleton rest singleton
func ResetSingleton(options ...storage.Option) {
	current := newManager(options...)
	singletonMu.Lock()
	singleton = current
	singletonMu.Unlock()
}
