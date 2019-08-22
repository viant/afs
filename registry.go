package afs

import (
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"sync"
)

//Provider represents manager provider
type Provider func(options ...storage.Option) (storage.Manager, error)

//Registry represents  abstract file system service provider registry
type Registry interface {
	//Register register schemeURL with storage service
	Register(uRLScheme string, provider Provider)

	//Get returns service provider for supplied schemeURL
	Get(uRLScheme string) (Provider, error)
}

type registry struct {
	proviers map[string]Provider
	*sync.RWMutex
}

func (r *registry) Register(uRLScheme string, provider Provider) {
	r.Lock()
	defer r.Unlock()
	r.proviers[uRLScheme] = provider

}

func (r *registry) Get(uRLScheme string) (Provider, error) {
	r.RLock()
	defer r.RUnlock()
	provider, ok := r.proviers[uRLScheme]
	if !ok {
		return nil, fmt.Errorf("failed to lookup storege provider %v", uRLScheme)
	}
	return provider, nil
}

var singleton Registry

//GetRegistry return singleton registry
func GetRegistry() Registry {
	if singleton != nil {
		return singleton
	}
	singleton = &registry{
		proviers: make(map[string]Provider),
		RWMutex:  &sync.RWMutex{},
	}
	return singleton
}

//Manager returns a manager for supplied sourceURL
func Manager(URL string, options ...storage.Option) (storage.Manager, error) {
	scheme := url.Scheme(URL, file.Scheme)
	provider, err := GetRegistry().Get(scheme)
	if err != nil {
		return nil, err
	}
	return provider(options)
}
