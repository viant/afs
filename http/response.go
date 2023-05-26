package http

import (
	"net/http"
)

func (s *manager) closeResponse(response *http.Response) {
	if response.Body != nil {
		_ = response.Body.Close()
	}

}
