package scp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
)

type manager struct {
	*base.Manager
}

func (m *manager) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	_, URLPath := url.Base(URL, Scheme)
	srv, err := m.Storager(ctx, URL, options)
	if err != nil {
		return nil, nil, err
	}
	service, ok := srv.(*storager)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported storager type: expected: %T, but had %T", service, srv)
	}
	return service.Uploader(ctx, URLPath)
}

func (m *manager) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	baseURL, URLPath := url.Base(URL, Scheme)
	match, modifier := option.GetWalkOptions(options)
	srv, err := m.Storager(ctx, baseURL, options)
	if err != nil {
		return err
	}
	service, ok := srv.(*storager)
	if !ok {
		return fmt.Errorf("unsupported storager type: expected: %T, but had %T", service, srv)
	}
	return service.Walk(ctx, URLPath, func(parent string, info os.FileInfo, reader io.Reader) (shallContinue bool, err error) {
		if !match(parent, info) {
			return true, nil
		}
		readerCloser := ioutil.NopCloser(reader)
		if modifier != nil {
			if info, readerCloser, err = modifier(info, readerCloser); err != nil {
				return false, err
			}
		}
		shallContinue, err = handler(ctx, baseURL, parent, info, readerCloser)
		return shallContinue, err
	})

}

func (m *manager) provider(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	options = m.Options(options)
	timeout := option.Timeout{}
	var basicAuth option.BasicAuth
	var keyAuth KeyAuth
	var authProvider AuthProvider
	option.Assign(options, &basicAuth, &keyAuth, &authProvider, &timeout)
	if timeout.Duration == 0 {
		timeout = option.NewTimeout(defaultTimeoutMs)
	}
	if basicAuth == nil && keyAuth == nil && authProvider == nil {
		keyAuth, _ = LocalhostKeyAuth("")
	}
	if authProvider == nil {
		authProvider = NewAuthProvider(keyAuth, basicAuth)
	}
	config, err := authProvider.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ssh config")
	}
	host := url.Host(baseURL)
	return NewStorager(host, timeout.Duration, config)
}

func newManager(options ...storage.Option) *manager {
	result := &manager{}
	baseMgr := base.New(result, Scheme, result.provider, options)
	result.Manager = baseMgr
	return result
}

//New creates scp manager
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}
