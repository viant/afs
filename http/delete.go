package http

import (
	"context"
	"fmt"
	"github.com/viant/afs/storage"
	"net/http"
)

//Delete sends delete method with supplied URL
func (s *manager) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	request, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return err
	}
	response, err := s.run(ctx, URL, request, options...)
	if err != nil {
		return err
	}
	defer s.closeResponse(response)
	if IsStatusOK(response) {
		return nil
	}
	return fmt.Errorf("invalid status code: %v", response.StatusCode)
}
