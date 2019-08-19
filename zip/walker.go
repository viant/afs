package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"io"
	"io/ioutil"
	"path"
)

type walker struct {
	storage.Downloader
}

func (w *walker) Walk(ctx context.Context, URL string, handler storage.OnVisit, options ...storage.Option) error {
	URL = url.Normalize(URL, file.Scheme)
	rawReader, err := w.DownloadWithURL(ctx, URL, options...)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(rawReader)
	if err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
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
		buffer.Reset()

	}
	return nil
}

//NewWalker returns a walker
func NewWalker(download storage.Downloader) storage.Walker {
	return &walker{Downloader: download}
}
