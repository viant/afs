package http

import (
	"context"
	"fmt"
	"github.com/viant/afs/storage"
	"io"
	"net/http"
	"os"
)

//Upload sends put request to supplied URL with provided reader
func (s *manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	request, err := http.NewRequest(http.MethodPut, URL, reader)
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
