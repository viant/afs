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
)

func TestManager_Exists(t *testing.T) {
	testPort := 8873
	baseURL := fmt.Sprintf("http://localhost:%v", testPort)

	var useCases = []struct {
		description string
		URL         string
		getParrot   *http.Response
		headParrot  *http.Response
		exists      bool
		hasError    bool
	}{
		{
			description: "head based exists",
			URL:         url.Join(baseURL, "/foo/asset1.txt"),
			headParrot: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("test")),
			},
			exists: true,
		},
		{
			description: "get based exists",
			URL:         url.Join(baseURL, "/foo/asset2.txt"),
			getParrot: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("test")),
			},
			exists: true,
		},
		{
			description: "not found",
			URL:         url.Join(baseURL, "/foo/asset3.txt"),
		},
		{
			description: "http error",
			URL:         url.Join("http://localhost:2222", "/foo/asset4.txt"),
			hasError:    true,
		},
	}

	ctx := context.Background()
	parrots := map[string]*http.Response{}

	for _, useCase := range useCases {
		if useCase.getParrot != nil {
			addGetURLParrots(useCase.URL, useCase.getParrot, parrots)
		}
		if useCase.headParrot != nil {
			addHeadURLParrots(useCase.URL, useCase.headParrot, parrots)
		}
	}
	go startServer(testPort, parrotHandler(parrots))

	for _, useCase := range useCases {
		manager := newManager()
		actual, err := manager.Exists(ctx, useCase.URL)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.exists, actual)
	}

}
