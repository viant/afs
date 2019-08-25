package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/archive"
	"github.com/viant/afs/asset"
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
	//underlying archive URL
	walker     *walker
	URL        string
	exists     bool
	closer     io.Closer
	uploader   storage.Uploader
	downloader storage.Downloader
}

//Exists returns true if resource exists in archive
func (s *storager) Exists(ctx context.Context, location string) (bool, error) {
	objects, _ := s.List(ctx, location)
	return len(objects) > 0, nil
}

//List lists archive assets
func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	if !s.exists {
		return nil, fmt.Errorf("%v: not found", location)
	}
	var result = make([]os.FileInfo, 0)
	location = strings.Trim(location, "/")
	basicMatcher, _ := matcher.NewBasic(location, "", "")
	var listMatcher option.Matcher
	option.Assign(options, &listMatcher)
	listMatcher = option.GetMatcher(listMatcher)
	err := s.walker.Walk(ctx, s.URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if !basicMatcher.Match(parent, info) {
			return true, nil
		}
		if !listMatcher(parent, info) {
			return true, nil
		}
		result = append(result, info)
		return true, nil
	})
	return result, err
}

//Walk visits location resources
func (s *storager) Walk(ctx context.Context, location string, handler func(parent string, info os.FileInfo, reader io.Reader) (bool, error)) error {
	if !s.exists {
		return fmt.Errorf("%v: not found", location)
	}
	location = strings.Trim(location, "/")
	basicMatcher, _ := matcher.NewBasic(location, "", "")
	return s.walker.Walk(ctx, s.URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if !basicMatcher.Match(parent, info) {
			return true, nil
		}
		return handler(parent, info, reader)
	})
}

//Download fetches content for supplied location
func (s *storager) Download(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	if !s.exists {
		return nil, fmt.Errorf("%v: not found", location)
	}
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
		return nil, fmt.Errorf("%v: not found", location)
	}
	return result, err
}

//Delete removes specified resource from archive
func (s *storager) Delete(ctx context.Context, location string) error {
	if !s.exists {
		return fmt.Errorf("%v: not found", location)
	}
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
	return s.uploader.Upload(ctx, s.URL, file.DefaultFileOsMode, uploader.buffer)
}

func (s *storager) touch(ctx context.Context) error {
	buffer := new(bytes.Buffer)
	writer := zip.NewWriter(buffer)
	_ = writer.Flush()
	_ = writer.Close()
	err := s.uploader.Upload(ctx, s.URL, file.DefaultFileOsMode, buffer)
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
	uploader := archive.NewRewriteUploader(func(resources []*asset.Resource) error {
		uploader := newBatchUploader(nil)
		upload, closer, err := uploader.Uploader(ctx, "")
		if err != nil {
			return errors.Wrapf(err, "failed to upload: %v in archvive: %v", destination, s.URL)
		}
		err = archive.Rewrite(ctx, s.walker, s.URL, upload, archive.UploadHandler(resources))
		if err == nil {
			err = closer.Close()
		}
		if err != nil {
			return err
		}
		s.walker.data = uploader.buffer.Bytes()
		return s.uploader.Upload(ctx, s.URL, file.DefaultFileOsMode, uploader.buffer)
	})
	return uploader.Upload, uploader, nil
}

//Upload uploads content for supplied destination, if archive does not exists, it creates one
func (s *storager) Upload(ctx context.Context, destination string, mode os.FileMode, content []byte, options ...storage.Option) error {
	return s.Create(ctx, destination, mode, content, false)
}

//Create creates a file or directory in archive, if archive does not exists, it creates one
func (s *storager) Create(ctx context.Context, destination string, mode os.FileMode, content []byte, isDir bool, options ...storage.Option) error {
	if !s.exists {
		if err := s.touch(ctx); err != nil {
			return err
		}
	}
	uploader := newBatchUploader(nil)
	upload, closer, err := uploader.Uploader(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to create: %v in archive: %v", destination, s.URL)
	}
	err = archive.Rewrite(ctx, s.walker, s.URL, upload, archive.CreateHandler(destination, mode, content, isDir))
	if err == nil {
		err = closer.Close()
	}
	if err != nil {
		return err
	}
	s.walker.data = uploader.buffer.Bytes()
	return s.uploader.Upload(ctx, s.URL, file.DefaultFileOsMode, uploader.buffer)
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
	object, _ := mgr.List(ctx, URL)
	if len(object) == 1 && object[0].IsDir() {
		return nil, fmt.Errorf("%v: is directory", URL)
	}
	return &storager{
		walker:     newWalker(mgr),
		exists:     len(object) == 1,
		closer:     mgr,
		uploader:   mgr,
		downloader: mgr,
		URL:        URL,
	}, nil
}

//NewStorager create a storage service
func NewStorager(ctx context.Context, baseURL string, mgr storage.Manager) (storage.Storager, error) {
	return newStorager(ctx, baseURL, mgr)
}
