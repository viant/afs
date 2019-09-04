package http

import (
	"github.com/viant/afs/storage"
	"net/http"
)

type manager struct {
	client         *http.Client
	baseURLClients map[string]*http.Client
	options        []storage.Option
}

func CloseIdleConnections(client interface{}) {
	type closeIdler interface {
		CloseIdleConnections()
	}
	if closer, ok := client.(closeIdler); ok {
		closer.CloseIdleConnections()
	}
}

func (s *manager) Close() error {
	if s.client != nil {
		CloseIdleConnections(s.client)
	}
	for _, client := range s.baseURLClients {
		CloseIdleConnections(client)
	}
	return nil
}

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
