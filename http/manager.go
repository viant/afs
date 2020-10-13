package http

import (
	"github.com/viant/afs/storage"
	"net/http"
	"sync"
)

type manager struct {
	client         *http.Client
	mux            sync.Mutex
	baseURLClients map[string]*http.Client
	options        []storage.Option
}

//CloseIdleConnections closes iddle connections
func CloseIdleConnections(client interface{}) {
	type closeIdler interface {
		CloseIdleConnections()
	}
	if closer, ok := client.(closeIdler); ok {
		closer.CloseIdleConnections()
	}
}

//Close closes mananger
func (s *manager) Close() error {
	if s.client != nil {
		CloseIdleConnections(s.client)
	}
	for _, client := range s.baseURLClients {
		CloseIdleConnections(client)
	}
	return nil
}

//Scheme returns schmea
func (s *manager) Scheme() string {
	return Scheme
}

func newManager(options ...storage.Option) *manager {
	return &manager{
		options:        options,
		baseURLClients: make(map[string]*http.Client),
	}
}

//New creates HTTP manager
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}
