package http

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/storage"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestClientProvider(t *testing.T) {
	const testTimeout = 40000
	testError := false

	var testProvider = func(baseURL string, options ...storage.Option) (*http.Client, error) {
		roundTripper := http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   testTimeout,
				KeepAlive: 10000,
			}).DialContext}
		client := &http.Client{
			Transport: &roundTripper,
			Timeout:   testTimeout,
		}
		if testError {
			return nil, fmt.Errorf("test error")
		}
		return client, nil
	}

	var useCases = []struct {
		description   string
		baseURL       string
		clientTimeout time.Duration
		options       []storage.Option
		hasError      bool
	}{
		{
			description:   "custom client",
			baseURL:       "",
			clientTimeout: testTimeout,
			options: []storage.Option{
				testProvider,
			},
		},
		{
			description:   "default HTTP client",
			baseURL:       "/foo",
			clientTimeout: http.DefaultClient.Timeout,
		},
		{
			description: "test error",
			baseURL:     "/foo",
			hasError:    true,
			options: []storage.Option{
				testProvider,
			},
		},
	}

	for _, useCase := range useCases {
		manager := newManager(useCase.options...)
		testError = useCase.hasError
		client1, err := manager.getClient(useCase.baseURL)
		if testError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, useCase.clientTimeout, int(client1.Timeout), useCase.description)
		client2, _ := manager.getClient(useCase.baseURL)
		assert.Equal(t, client1, client2, useCase.description)
		_ = manager.Close()
	}
}
