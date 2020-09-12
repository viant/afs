package tar

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/archive"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/base"
	"github.com/viant/afs/file"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type storager struct {
	base.Storager
	//underlying archive URL
	walker     *walker
	mode       os.FileMode
	URL        string
	exists     bool
	closer     io.Closer
	uploader   storage.Uploader
	downloader storage.Opener
}

//Exists returns true if resource exists in archive
func (s *storager) Exists(ctx context.Context, location string, options ...storage.Option) (bool, error) {
	objects, _ := s.List(ctx, location)
	return len(objects) > 0, nil
}

//List lists archive assets
func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	if !s.exists {
		return nil, fmt.Errorf("%v: not found", s.URL)
	}
	var result = make([]os.FileInfo, 0)
	location = strings.Trim(location, "/")
	basicMatcher, _ := matcher.NewBasic(location, "", "", nil)
	match, page := option.GetListOptions(options)

	err := s.walker.Walk(ctx, s.URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if !basicMatcher.Match(parent, info) {
			return true, nil
		}
		if !match(parent, info) {
			return true, nil
		}
		page.Increment()
		if page.ShallSkip() {
			return true, nil
		}
		result = append(result, info)
		if page.HasReachedLimit() {
			return false, nil
		}
		return true, nil
	})
	return result, err
}

//Walk visits location resources
func (s *storager) Walk(ctx context.Context, location string, handler func(parent string, info os.FileInfo, reader io.Reader) (bool, error), options ...storage.Option) error {
	if !s.exists {
		return fmt.Errorf("%v: not found", s.URL)
	}
	location = strings.Trim(location, "/")
	basicMatcher, _ := matcher.NewBasic(location, "", "", nil)
	match, modifier := option.GetWalkOptions(options)
	return s.walker.Walk(ctx, s.URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if !basicMatcher.Match(parent, info) {
			return true, nil
		}

		if !match(parent, info) {
			return true, nil
		}
		if modifier != nil {
			info, reader, err = modifier(info, ioutil.NopCloser(reader))
			if err != nil {
				return false, err
			}
		}
		return handler(parent, info, reader)
	})
}

//Open fetches content for supplied location
func (s *storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	if !s.exists {
		return nil, fmt.Errorf("%v: not found", s.URL)
	}
	location = strings.Trim(location, "/")
	var result io.ReadCloser
	err := s.walker.Walk(ctx, s.URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		filename := path.Join(parent, info.Name())
		if location == filename {
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				return false, err
			}
			result = ioutil.NopCloser(bytes.NewReader(data))
			return false, nil
		}
		return true, nil
	})
	if err == nil && result == nil {
		return nil, fmt.Errorf("%v: not found in archive: %v", location, s.URL)
	}
	return result, err
}

//Delete removes specified resource from archive
func (s *storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	if !s.exists {
		return fmt.Errorf("%v: not found", s.URL)
	}
	location = strings.Trim(location, "/")
	uploader := newBatchUploader(nil)
	upload, closer, err := uploader.Uploader(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to delete: %v in archive: %v", location, s.URL)
	}

	err = archive.Rewrite(ctx, s.walker, s.URL, upload, archive.DeleteHandler(location))
	if err == nil {
		err = closer.Close()
	}
	if err != nil {
		return err
	}
	return s.uploader.Upload(ctx, s.URL, s.mode, uploader.buffer)
}

func (s *storager) touch(ctx context.Context) error {
	buffer := new(bytes.Buffer)
	writer := zip.NewWriter(buffer)
	_ = writer.Flush()
	_ = writer.Close()
	err := s.uploader.Upload(ctx, s.URL, s.mode, buffer)
	if err == nil {
		s.exists = true
	}
	return err
}

//Uploader return batch uploader, if archive does not exists, it creates one
func (s *storager) Uploader(ctx context.Context, destination string) (storage.Upload, io.Closer, error) {
	if !s.exists {
		if err := s.touch(ctx); err != nil {
			return nil, nil, err
		}
	}
	destination = strings.Trim(destination, "/")
	uploader := archive.NewRewriteUploader(func(resources []*asset.Resource) error {
		uploader := newBatchUploader(nil)
		upload, closer, err := uploader.Uploader(ctx, "")
		if err != nil {
			return errors.Wrapf(err, "failed to upload: %v in archive: %v", destination, s.URL)
		}
		resources = archive.UpdateDestination(destination, resources)
		err = archive.Rewrite(ctx, s.walker, s.URL, upload, archive.UploadHandler(resources))
		if err == nil {
			err = closer.Close()
		}
		if err != nil {
			return err
		}
		s.walker.data = uploader.buffer.Bytes()
		return s.uploader.Upload(ctx, s.URL, s.mode, uploader.buffer)
	})
	return uploader.Upload, uploader, nil
}

//Upload uploads content for supplied destination, if archive does not exists, it creates one
func (s *storager) Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	return s.Create(ctx, destination, mode, reader, false)
}

//Create creates a file or directory in archive, if archive does not exists, it creates one
func (s *storager) Create(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, isDir bool, options ...storage.Option) error {
	if !s.exists {
		if err := s.touch(ctx); err != nil {
			return err
		}
	}
	destination = strings.Trim(destination, "/")
	uploader := newBatchUploader(nil)
	upload, closer, err := uploader.Uploader(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to create: %v in archive: %v", destination, s.URL)
	}

	var content []byte
	if reader != nil {
		if content, err = ioutil.ReadAll(reader); err != nil {
			return err
		}
	}

	err = archive.Rewrite(ctx, s.walker, s.URL, upload, archive.CreateHandler(destination, mode, content, isDir))
	if err == nil {
		err = closer.Close()
	}
	if err != nil {
		return err
	}
	s.walker.data = uploader.buffer.Bytes()
	return s.uploader.Upload(ctx, s.URL, s.mode, uploader.buffer)
}

//Close closes undelrying closer
func (s *storager) Close() error {
	return s.closer.Close()
}

//newStorager create a storage service
func newStorager(ctx context.Context, baseURL string, mgr storage.Manager) (*storager, error) {
	URL := url.SchemeExtensionURL(baseURL)
	if URL == "" {
		return nil, fmt.Errorf("invalid URL: %v", baseURL)
	}
	mode := file.DefaultFileOsMode
	object, _ := mgr.List(ctx, URL)
	if len(object) == 1 {
		if object[0].IsDir() {
			return nil, fmt.Errorf("%v: is directory", URL)
		}
		mode = object[0].Mode()
	}
	result := &storager{
		walker:     newWalker(mgr),
		exists:     len(object) == 1,
		closer:     mgr,
		mode:       mode,
		uploader:   mgr,
		downloader: mgr,
		URL:        URL,
	}
	result.Storager.List = result.List
	return result, nil
}

//NewStorager create a storage service
func NewStorager(ctx context.Context, baseURL string, mgr storage.Manager) (storage.Storager, error) {
	return newStorager(ctx, baseURL, mgr)
}
