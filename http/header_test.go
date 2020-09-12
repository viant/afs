package http

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/url"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {

	testPort := 8879
	baseURL := fmt.Sprintf("http://localhost:%v", testPort)
	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		expect      string
		putParrot   *http.Response
		header      http.Header
		hasError    bool
	}{
		{
			description: "asset delete",
			URL:         url.Join(baseURL, "/foo/bar.txt"),
			expect:      "test is test",
			header: http.Header{
				"Set-Cookie": []string{
					"id=a3fWa; Expires=Wed, 21 Oct 2035 07:28:00 GMT; Secure; HttpOnly",
				},
			},
			putParrot: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("test is test")),
			},
		},
		{
			description: "not found error delete",
			URL:         url.Join(baseURL, "/foo/error.txt"),
			hasError:    true,
		},
	}

	parrots := map[string]*http.Response{}
	for _, useCase := range useCases {
		addURLParrots(http.MethodGet, useCase.URL, useCase.putParrot, parrots)
	}
	go startServer(testPort, parrotHandler(parrots))

	for _, useCase := range useCases {
		manager := newManager()
		reader, err := manager.OpenURL(ctx, useCase.URL, useCase.header)
		if err != nil && useCase.hasError {
			assert.NotNil(t, err, useCase.description)
		}
		if useCase.hasError {
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		data, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.EqualValues(t, useCase.expect, string(data))

	}
}

func TestHeaderTime(t *testing.T) {

	var useCases = []struct {
		description string
		key         string
		value       string
		expect      string
	}{

		{
			description: "exact header",
			key:         "Last-Modified",
			value:       "Tue, 20 Oct 2019 01:02:20 GMT",
		},
		{
			description: "mix case  header",
			key:         "last-Modified",
			value:       "Tue, 20 Oct 2019 01:02:20 GMT",
		},
		{
			description: "mix case  header",
			key:         "last-ssodified",
		},
		{
			description: "empty",
			key:         "",
		},
	}

	for _, useCase := range useCases {
		header := http.Header{}
		if useCase.key != "" {
			header.Set(useCase.key, useCase.value)
		}

		now := time.Now()
		actual := HeaderTime(header, useCase.key, now)
		if useCase.value != "" {
			expectModified, err := ParseHTTPDate(useCase.value)
			assert.Nil(t, err, useCase.description)
			assert.EqualValues(t, expectModified, actual, useCase.description)
		} else {
			assert.EqualValues(t, now, actual, useCase.description)
		}

	}

}
