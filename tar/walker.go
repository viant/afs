package tar

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type walker struct {
	storage.Downloader
	data []byte
	URL  string
}

func (w *walker) load(ctx context.Context, URL string, options ...storage.Option) ([]byte, error) {
	if len(w.data) > 0 && URL == w.URL {
		return w.data, nil
	}
	rawReader, err := w.DownloadWithURL(ctx, URL, options...)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(rawReader)
	if err != nil {
		return nil, err
	}
	w.URL = URL
	w.data = data
	return w.data, nil
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
	data, err := w.load(ctx, URL, options...)
	if err != nil {
		return err
	}
	reader := tar.NewReader(bytes.NewReader(data))
	buffer := new(bytes.Buffer)
	//cache is only used if sym link are used
	var cache = make(map[string][]byte)
outer:
	for {
		header, err := reader.Next()
		if err == io.EOF || header == nil {
			break
		}
		relative, name := path.Split(header.Name)
		mode := header.Mode
		if header.Typeflag == tar.TypeSymlink {
			mode |= int64(os.ModeSymlink)
		}
		if header.Typeflag == tar.TypeDir {
			mode |= int64(os.ModeDir)
		}

		info := file.NewInfo(name, header.Size, os.FileMode(mode), header.ModTime, header.Typeflag == tar.TypeDir)

		switch header.Typeflag {
		case tar.TypeDir:
			shallContinue, err := handler(ctx, URL, relative, info, nil)
			if err != nil || !shallContinue {
				break outer
			}

		case tar.TypeReg:
			_, err = io.Copy(buffer, reader)
			if err != nil {
				return err
			}
			shallContinue, err := handler(ctx, URL, relative, info, buffer)
			if err != nil || !shallContinue {
				break outer
			}
			buffer.Reset()
		case tar.TypeSymlink:
			linkPath := path.Clean(path.Join(relative, header.Linkname))
			reader, err := w.fetch(tar.NewReader(bytes.NewReader(data)), linkPath, cache)
			if err != nil {
				return err
			}
			link := object.NewLink(header.Linkname, url.Join(URL, linkPath), nil)
			info = file.NewInfo(name, header.Size, os.FileMode(mode), header.ModTime, header.Typeflag == tar.TypeDir, link)
			_, err = io.Copy(buffer, reader)
			if err != nil {
				return err
			}
			shallContinue, err := handler(ctx, URL, relative, info, buffer)
			if err != nil || !shallContinue {
				return err
			}
			buffer.Reset()
		default:
			return fmt.Errorf("unknown header type: %v", header.Typeflag)
		}
	}
	return nil
}

//newWalker returns a walker
func newWalker(download storage.Downloader) *walker {
	return &walker{Downloader: download}
}

//NewWalker returns a walker
func NewWalker(download storage.Downloader) storage.Walker {
	return newWalker(download)
}
