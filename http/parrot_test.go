package http

import (
	"fmt"
	"github.com/viant/afs/url"
	"io/ioutil"
	"net/http"
)

func addGetURLParrots(URL string, response *http.Response, result map[string]*http.Response) {
	addURLParrots(http.MethodGet, URL, response, result)
}

func addHeadURLParrots(URL string, response *http.Response, result map[string]*http.Response) {
	addURLParrots(http.MethodHead, URL, response, result)
}

func addURLParrots(method, URL string, response *http.Response, result map[string]*http.Response) {
	if response == nil {
		return
	}
	_, URLPath := url.Base(URL, Scheme)
	key := method + ":" + URLPath
	result[key] = response
}

func parrotHandler(responses map[string]*http.Response) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		key := fmt.Sprintf("%v:%v", request.Method, request.URL.Path)
		response, ok := responses[key]

		if !ok {
			http.NotFound(writer, request)
			return
		}

		if len(response.Header) > 0 {
			for k, v := range response.Header {
				writer.Header().Set(k, v[0])
			}
		}
		if data, err := ioutil.ReadAll(response.Body); err == nil {
			_, _ = writer.Write(data)
		}
	}
}
