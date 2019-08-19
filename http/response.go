package http

import (
	"io/ioutil"
	"net/http"
)

func (s *manager) closeResponse(response *http.Response) {
	if response.Body != nil {
		_, _ = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
	}

}
