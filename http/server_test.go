package http

import (
	"fmt"
	"net/http"
)

func startServer(port int, handler func(writer http.ResponseWriter, request *http.Request)) {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: newHandler(handler)}
	_ = httpServer.ListenAndServe()
}

type handler struct {
	onRequest func(writer http.ResponseWriter, request *http.Request)
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.onRequest(writer, request)
}

func newHandler(onRequest func(writer http.ResponseWriter, request *http.Request)) http.Handler {
	return &handler{onRequest: onRequest}
}
