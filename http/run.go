package http

import (
	"context"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"net/http"
)

func (s *manager) run(ctx context.Context, URL string, request *http.Request, options ...storage.Option) (*http.Response, error) {
	var clientProvider ClientProvider
	var reader io.Reader
	var basicAuthProvider option.BasicAuth
	resp := &http.Response{}
	header := http.Header{}
	option.Assign(options, &clientProvider, &basicAuthProvider, &header, &reader, &resp)
	s.setHeader(request, header)
	s.authWithBasicCred(request, basicAuthProvider)
	client, err := s.getClient(URL, options...)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		request.WithContext(ctx)
	}
	response, err := client.Do(request)
	if err == nil && resp != nil {
		*resp = *response
	}
	return response, err
}
