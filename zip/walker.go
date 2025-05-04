package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
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

// Walk walks the zip file and calls the handler for each file and directory.
// This is inconsistent with the expected behavior of other List() operations.
// List() for a file system requires the recursive flag, and in the recursive case, does not include the root directory.
// This function will always be recursive, and includes the root directory.
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

	processedDirs := make(map[string]bool)

	// cache is only used if sym link are used
	for _, fileHandler := range reader.File {
		parentPath, name := path.Split(fileHandler.Name)
		if parentPath != "" {
			// Split() returns a trailing slash
			parentPath = parentPath[:len(parentPath)-1]
		}

		fileInfo := fileHandler.FileInfo()
		fileModTime := fileInfo.ModTime()

		dirPath := parentPath

		// For files, process parent directories first (implicit directories)
		// Note that the root directory has dirPath of "" (instead of "." or "/").
		// This is because of the for condition and behavior of path.Dir()
		for !fileHandler.Mode().IsDir() && dirPath != "." && dirPath != "/" {
			if !processedDirs[dirPath] {
				// markPath is the path to be used to mark the directory as processed
				var markPath string

				var parentDir string

				// special case for root, we need to establish a fake initial dir
				if dirPath == "" {
					markPath = ""
					parentDir = ""
				} else {
					markPath = dirPath
					parentDir = path.Dir(dirPath)
				}

				// Create virtual directory entry.
				// There is a use case of List() where there is a prefix filter, which breaks if the directory FileInfo has a name.
				// However, this results in a bunch of pretty much identical directory FileInfos.
				// For Copy(), this will cause a bunch of repeat directories to be created.
				dirInfo := file.NewInfo("", 0, os.ModeDir|0755, fileModTime, true)

				// Visit the directory
				shallContinue, err := handler(ctx, URL, parentDir, dirInfo, nil)
				if err != nil || !shallContinue {
					return err
				}

				processedDirs[markPath] = true
			}

			dirPath = path.Dir(dirPath)
		}

		// note that in the case of a directory, the name is always "" due to path.Split()
		info := file.NewInfo(name,
			fileInfo.Size(), fileInfo.Mode(), fileModTime, fileInfo.IsDir())

		var reader io.ReadCloser
		if !fileHandler.Mode().IsDir() {
			if reader, err = fileHandler.Open(); err != nil {
				return err
			}
		} else {
			// handle case if the entry is a directory
			processedDirs[fileHandler.Name] = true
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

// NewWalker returns a walker
func newWalker(download storage.Opener) *walker {
	return &walker{Opener: download}
}

// NewWalker returns a walker
func NewWalker(downloader storage.Opener) storage.Walker {
	return newWalker(downloader)
}
