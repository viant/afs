package http

import (
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"net/http"
)

//ClientProvider represents clinet provider option
type ClientProvider func(baseURL string, options ...storage.Option) (*http.Client, error)

func (s *manager) getClient(baseURL string, options ...storage.Option) (*http.Client, error) {
	baseURL, _ = url.Base(baseURL, Scheme)
	s.mux.Lock()
	defer s.mux.Unlock()
	client, ok := s.baseURLClients[baseURL]
	if ok {
		return client, nil
	}
	if len(s.options) > 0 {
		options = append(s.options, options...)
	}
	var clientProvider ClientProvider
	option.Assign(options, &clientProvider)
	if clientProvider == nil {
		if s.client == nil {
			s.client = http.DefaultClient
		}
		return s.client, nil
	}
	var err error
	if clientProvider != nil {
		if client, err = clientProvider(baseURL, options); err != nil {
			return nil, err
		}
		s.baseURLClients[baseURL] = client
	}
	return client, nil
}

func (s *manager) authWithBasicCred(request *http.Request, authenticator option.BasicAuth) {
	if authenticator == nil {
		return
	}
	username, password := authenticator.Credentials()
	request.SetBasicAuth(username, password)
}
