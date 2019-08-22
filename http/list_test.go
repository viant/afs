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

func TestManager_List(t *testing.T) {
	testPort := 8872
	baseURL := fmt.Sprintf("http://localhost:%v", testPort)

	var useCases = []struct {
		description  string
		URL          string
		lastModified string
		parrot       *http.Response
		expectSize   int
		hasError     bool
	}{
		{
			description:  "asset test with modified header",
			URL:          url.Join(baseURL, "/foo/asset1.txt"),
			lastModified: "Tue, 20 Oct 2019 01:02:20 GMT",
			parrot: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("test")),
			},
			expectSize: 4,
		},
		{
			description: "asset not found error",
			hasError:    true,
			URL:         url.Join(baseURL, "/foo/bar.txt"),
		},
	}

	ctx := context.Background()
	parrots := map[string]*http.Response{}

	for _, useCase := range useCases {
		if useCase.lastModified != "" {
			if len(useCase.parrot.Header) == 0 {
				useCase.parrot.Header = http.Header{}
			}
			useCase.parrot.Header.Set(lastModifiedHeader, useCase.lastModified)
		}
		addGetURLParrots(useCase.URL, useCase.parrot, parrots)
	}
	go startServer(testPort, parrotHandler(parrots))

	for _, useCase := range useCases {
		manager := New()
		objects, err := manager.List(ctx, useCase.URL)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.Equal(t, 1, len(objects), useCase.description)
		assert.Equal(t, useCase.URL, objects[0].URL(), useCase.description)
		assert.EqualValues(t, useCase.expectSize, objects[0].Size(), useCase.description)

		if useCase.lastModified != "" {
			expectModified, err := ParseHTTPDate(useCase.lastModified)
			assert.Nil(t, err, useCase.description)
			assert.Equal(t, expectModified, objects[0].ModTime(), useCase.description)
		}
	}

}
