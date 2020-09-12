package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"path"
)

type walker struct {
	storage.Opener
	data []byte
	URL  string
}

func (w *walker) open(ctx context.Context, URL string, options ...storage.Option) (io.ReaderAt, int, error) {
	if len(w.data) > 0 && URL == w.URL {
		return bytes.NewReader(w.data), len(w.data), nil
	}
	rawReader, err := w.OpenURL(ctx, URL, options...)
	if err != nil {
		return nil, 0, err
	}
	size := option.Size(0)
	option.Assign(options, &size)
	if readerAt, ok := rawReader.(io.ReaderAt); ok && size > 0 {
		return readerAt, int(size), nil
	}
	defer rawReader.Close()
	data, err := ioutil.ReadAll(rawReader)
	if err != nil {
		return nil, 0, err
	}
	w.URL = URL
	w.data = data
	return bytes.NewReader(w.data), len(w.data), nil
}

func (w *walker) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	readerAt, size, err := w.open(ctx, URL, options...)
	if err != nil {
		return err
	}
	if closer, ok := readerAt.(io.Closer); ok {
		defer closer.Close()
	}
	reader, err := zip.NewReader(readerAt, int64(size))
	if err != nil {
		return err
	}
	//cache is only used if sym link are used
	for _, fileHandler := range reader.File {
		parentPath, name := path.Split(fileHandler.Name)
		fileInfo := fileHandler.FileInfo()
		info := file.NewInfo(name, fileInfo.Size(), fileInfo.Mode(), fileInfo.ModTime(), fileInfo.IsDir())
		var reader io.ReadCloser
		if !fileHandler.Mode().IsDir() {
			if reader, err = fileHandler.Open(); err != nil {
				return err
			}
		}
		shallContinue, err := handler(ctx, URL, parentPath, info, reader)
		if reader != nil {
			err = reader.Close()
		}
		if err != nil || !shallContinue {
			return err
		}
	}
	return nil
}

//NewWalker returns a walker
func newWalker(download storage.Opener) *walker {
	return &walker{Opener: download}
}

//NewWalker returns a walker
func NewWalker(downloader storage.Opener) storage.Walker {
	return newWalker(downloader)
}
