package http

import (
	"net/http"
	"strings"
	"time"
)

func (s *manager) setHeader(request *http.Request, header http.Header) {
	if len(header) == 0 {
		return
	}
	if len(request.Header) == 0 {
		request.Header = header
	}
	for k, v := range header {
		request.Header[k] = v
	}
}

var timeLayouts = []string{"Mon, 02 Jan 2006 15:04:05 GMT", time.RFC850, time.ANSIC}

//HeaderTime returns time for header key
func HeaderTime(header http.Header, key string, defaultValue time.Time) time.Time {
	if len(header) == 0 {
		return defaultValue
	}
	value, ok := header[key]
	if !ok {
		key = strings.ToLower(key)
		for k, v := range header {
			if strings.ToLower(k) == key {
				value = v
			}
		}
	}

	if len(value) == 0 {
		return defaultValue
	}
	if result, err := ParseHTTPDate(value[0]); err == nil {
		return result
	}
	return defaultValue
}

//ParseHTTPDate parses date assigned
func ParseHTTPDate(value string) (result time.Time, err error) {
	for i := range timeLayouts {
		if result, err = time.Parse(timeLayouts[i], value); err == nil {
			return result, nil
		}
	}
	return result, err
}
