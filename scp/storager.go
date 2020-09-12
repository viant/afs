package scp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/base"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type storager struct {
	base.Storager
	address string
	*ssh.ClientConfig
	*ssh.Client
	timeout time.Duration
}

func (s *storager) connect() (err error) {
	if s.Client, err = ssh.Dial("tcp", s.address, s.ClientConfig); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to dial %s", s.address))
	}
	return nil
}

//Delete removes supplied asset
func (s *storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	session, err := s.NewSession()
	if err == nil {
		_, err = session.Output(fmt.Sprintf("rm -rf %v", location))
	}
	return err

}

//Exists returns true if location exists
func (s *storager) Exists(ctx context.Context, location string, options ...storage.Option) (bool, error) {
	session, err := newSession(s.Client, modeRead, true, s.timeout)
	if err != nil {
		return false, err
	}
	location = path.Clean(location)
	has := false
	_ = session.download(ctx, false, location, func(parent string, info os.FileInfo, reader io.Reader) (b bool, e error) {
		has = true
		return false, nil
	})
	return has, nil
}

//List lists location assets
func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	match, page := option.GetListOptions(options)
	var result = make([]os.FileInfo, 0)
	err := s.walk(ctx, location, false, func(relative string, info os.FileInfo, reader io.Reader) (shaleContinue bool, err error) {

		if !match(relative, info) {
			return true, nil
		}
		page.Increment()
		if page.ShallSkip() {
			return true, nil
		}
		result = append(result, info)
		return !page.HasReachedLimit(), nil
	})

	return result, err
}

//Walk visits location resources
func (s *storager) Walk(ctx context.Context, location string, handler func(relative string, info os.FileInfo, reader io.Reader) (bool, error), options ...storage.Option) error {
	return s.walk(ctx, location, true, handler)
}

//Walk visits location resources
func (s *storager) walk(ctx context.Context, location string, skipBaseLocation bool, handler func(relative string, info os.FileInfo, reader io.Reader) (bool, error), options ...storage.Option) error {
	session, err := newSession(s.Client, modeRead, true, s.timeout)
	if err != nil {
		return err
	}
	location = path.Clean(location)
	return session.download(ctx, skipBaseLocation, location, handler)
}

//Open fetches content for supplied location
func (s *storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	result := new(bytes.Buffer)
	err := s.Walk(ctx, location, func(relative string, info os.FileInfo, reader io.Reader) (b bool, e error) {
		_, err := io.Copy(result, reader)
		return false, err
	})
	return ioutil.NopCloser(result), err
}

//Uploader return batch uploader
func (s *storager) Uploader(ctx context.Context, destination string) (storage.Upload, io.Closer, error) {
	session, err := newSession(s.Client, modeWrite, true, 0)
	if err != nil {
		return nil, nil, err
	}
	return session.upload(destination)
}

//Upload uploads content for supplied destination
func (s *storager) Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	return s.Create(ctx, destination, mode, reader, false)
}

//Create creates a file or directory
func (s *storager) Create(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, isDir bool, options ...storage.Option) error {
	parent, name := path.Split(destination)
	if isDir {
		if session, err := s.NewSession(); err == nil {
			if _, err := session.Output(fmt.Sprintf("mkdir -p %s", destination)); err == nil {
				return nil
			}
		}
	}
	upload, closer, err := s.Uploader(ctx, parent)
	if err != nil {
		return err
	}
	defer func() { _ = closer.Close() }()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	info := file.NewInfo(name, int64(len(content)), mode, time.Now(), isDir)
	return upload(ctx, "", info, bytes.NewReader(content))
}

//FilterAuthOptions filters auth options
func filterAuthOption(options []storage.Option) (*ssh.ClientConfig, error) {
	var basicAuth option.BasicAuth
	var keyAuth KeyAuth
	var authProvider AuthProvider
	option.Assign(options, &basicAuth, &keyAuth, &authProvider)
	if basicAuth == nil && keyAuth == nil && authProvider == nil {
		return nil, nil
	}
	if authProvider == nil {
		authProvider = NewAuthProvider(keyAuth, basicAuth)
	}
	return authProvider.ClientConfig()
}

//IsAuthChanged return true if auth has changes
func (s *storager) IsAuthChanged(authOptions []storage.Option) bool {
	changed := s.isAuthChanged(authOptions)
	return changed
}

//IsAuthChanged return true if auth has changes
func (s *storager) isAuthChanged(authOptions []storage.Option) bool {
	if len(authOptions) == 0 {
		return false
	}
	sshConfig, _ := filterAuthOption(authOptions)
	if sshConfig == nil {
		return false
	}
	return sshConfig.User != s.ClientConfig.User
}

func (s *storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	options = append(options, option.NewPage(0, 1))
	objects, err := s.List(ctx, location, options)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, errors.Errorf("failed to get object: %v", location)
	}
	return objects[0], nil
}

//NewStorager returns a new storager
func NewStorager(address string, timeout time.Duration, config *ssh.ClientConfig) (storage.Storager, error) {
	if !strings.Contains(address, ":") {
		address += fmt.Sprintf(":%d", DefaultPort)
	}
	result := &storager{
		address:      address,
		ClientConfig: config,
		timeout:      timeout,
	}
	result.Storager.List = result.List

	return result, result.connect()
}
