package http

import "net/http"

//IsStatusOK returns true if status is 2xxx
func IsStatusOK(response *http.Response) bool {
	return response.StatusCode >= 200 && response.StatusCode <= 299
}
