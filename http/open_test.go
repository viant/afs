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

func TestManager_Download(t *testing.T) {

	testPort := 8875
	baseURL := fmt.Sprintf("http://localhost:%v", testPort)
	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		expect      string
		putParrot   *http.Response
		hasError    bool
	}{
		{
			description: "asset download",
			URL:         url.Join(baseURL, "/foo/bar.txt"),
			expect:      "test is test",
			putParrot: &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("test is test")),
			},
		},
		{
			description: "not found error download",
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
		reader, err := manager.OpenURL(ctx, useCase.URL)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		data, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		_ = reader.Close()
		assert.EqualValues(t, useCase.expect, string(data), useCase.description)

	}

}
