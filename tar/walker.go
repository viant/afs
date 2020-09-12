package tar

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type walker struct {
	storage.Opener
	data []byte
	URL  string
}

func (w *walker) open(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	if len(w.data) > 0 && URL == w.URL {
		return ioutil.NopCloser(bytes.NewReader(w.data)), nil
	}
	rawReader, err := w.OpenURL(ctx, URL, options...)
	if err != nil {
		return nil, err
	}

	size := option.Size(0)
	option.Assign(options, &size)
	if size > 0 {
		return rawReader, nil
	}
	defer rawReader.Close()
	data, err := ioutil.ReadAll(rawReader)
	if err != nil {
		return nil, err
	}
	w.URL = URL
	w.data = data
	return ioutil.NopCloser(bytes.NewReader(w.data)), nil
}

func (w *walker) fetch(reader *tar.Reader, location string, cache map[string][]byte) (io.Reader, error) {
	if len(cache) == 0 {
		w.buildCache(reader, cache)
	}
	data, ok := cache[location]
	if !ok {
		return nil, fmt.Errorf("%v: not found", location)
	}
	return bytes.NewReader(data), nil
}

func (w *walker) buildCache(reader *tar.Reader, cache map[string][]byte) {
	buffer := new(bytes.Buffer)
	for {
		header, err := reader.Next()
		if err == io.EOF || header == nil {
			break
		}

		if header.Typeflag == tar.TypeReg {
			_, _ = io.Copy(buffer, reader)
			if err != nil && err != io.EOF {
				break
			}
			copied := buffer.Bytes()
			dest := make([]byte, len(copied))
			copy(dest, copied)
			cache[header.Name] = dest
			buffer.Reset()
		}
	}
}

func (w *walker) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	readerCloser, err := w.open(ctx, URL, options...)
	if err != nil {
		return err
	}
	defer readerCloser.Close()
	var shallContinue bool
	var ioReader io.Reader
	reader := tar.NewReader(readerCloser)
	//cache is only used if sym link are used
	var cache = make(map[string][]byte)
	for {
		header, err := reader.Next()
		if err == io.EOF || header == nil {
			break
		}
		relative, name := path.Split(header.Name)
		mode := getFileMode(header)
		info := file.NewInfo(name, header.Size, os.FileMode(mode), header.ModTime, header.Typeflag == tar.TypeDir)
		switch header.Typeflag {
		case tar.TypeDir:
			shallContinue, err = handler(ctx, URL, relative, info, nil)
		case tar.TypeReg:
			shallContinue, err = visitRegularHeader(ctx, reader, handler, URL, relative, info)
		case tar.TypeSymlink:
			linkPath := path.Clean(path.Join(relative, header.Linkname))
			linkReader, err := w.open(ctx, URL, options...)
			if err != nil {
				return err
			}
			if ioReader, err = w.fetch(tar.NewReader(linkReader), linkPath, cache); err == nil {
				shallContinue, err = visitSymlinkHeader(ctx, header, linkPath, ioReader, handler, URL, relative, info)
			}
			linkReader.Close()
		default:
			return fmt.Errorf("unknown header type: %v", header.Typeflag)
		}
		if err != nil || !shallContinue {
			return err
		}
	}
	return nil
}

func getFileMode(header *tar.Header) int64 {
	mode := header.Mode
	if header.Typeflag == tar.TypeSymlink {
		mode |= int64(os.ModeSymlink)
	}
	if header.Typeflag == tar.TypeDir {
		mode |= int64(os.ModeDir)
	}
	return mode
}

func visitSymlinkHeader(ctx context.Context, header *tar.Header, linkPath string, reader io.Reader, handler storage.OnVisit, URL string, relative string, info os.FileInfo) (bool, error) {
	relative, name := path.Split(header.Name)
	link := object.NewLink(header.Linkname, url.Join(URL, linkPath), nil)
	info = file.NewInfo(name, header.Size, os.FileMode(info.Mode()), header.ModTime, header.Typeflag == tar.TypeDir, link)
	shallContinue, err := handler(ctx, URL, relative, info, reader)
	if err != nil || !shallContinue {
		return shallContinue, err
	}
	return true, nil
}

func visitRegularHeader(ctx context.Context, reader io.Reader, handler storage.OnVisit, URL string, relative string, info os.FileInfo) (bool, error) {
	shallContinue, err := handler(ctx, URL, relative, info, reader)
	if err != nil || !shallContinue {
		return shallContinue, err
	}
	return true, nil
}

//newWalker returns a walker
func newWalker(opener storage.Opener) *walker {
	return &walker{Opener: opener}
}

//NewWalker returns a walker
func NewWalker(download storage.Opener) storage.Walker {
	return newWalker(download)
}
