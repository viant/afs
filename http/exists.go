package http

import (
	"context"
	"github.com/viant/afs/storage"
	"net/http"
)

//Exists checks if asset exists
func (s *manager) Exists(ctx context.Context, URL string, options ...storage.Option) (bool, error) {

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPut} {
		request, err := http.NewRequest(method, URL, nil)
		if err != nil {
			return false, err
		}
		response, err := s.run(ctx, URL, request, options...)
		if err != nil {
			return false, err
		}
		if IsStatusOK(response) {
			return true, nil
		}
	}
	return false, nil
}
