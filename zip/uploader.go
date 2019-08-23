package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"

	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"os"
	"path"
)

type uploader struct {
	uploader storage.Uploader
}

func (u *uploader) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	var buffer *bytes.Buffer
	var uploader storage.Uploader
	option.Assign(options, &buffer, &uploader)
	if uploader == nil {
		uploader = u.uploader
	}
	if buffer == nil && uploader == nil {
		return nil, nil, fmt.Errorf("invalid options: %T and %T were empty", buffer, uploader)
	}
	if buffer == nil {
		buffer = new(bytes.Buffer)
	}
	writer := newWriter(ctx, buffer, URL, uploader)
	return func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error {
		filename := path.Join(parent, info.Name())
		mode := info.Mode().Perm()
		if info.IsDir() {
			mode |= os.ModeDir
		}
		info = file.NewInfo(filename, info.Size(), mode, info.ModTime(), info.IsDir())
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name = filename
		writer, err := writer.CreateHeader(header)
		if reader != nil {
			_, err = io.Copy(writer, reader)
			if err != nil {
				return err
			}
		}
		return err
	}, writer, nil
}

//NewBatchUploader returns a batch uploader
func NewBatchUploader(dest storage.Uploader) storage.BatchUploader {
	return &uploader{uploader: dest}
}
