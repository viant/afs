package tar

import (
	"archive/tar"
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
		link := ""
		var options []storage.Option
		if fileInfo, ok := info.(*file.Info); ok {
			link = fileInfo.Linkname
			options = make([]storage.Option, 0)
			options = append(options, fileInfo.Link)
		}
		filename := path.Join(parent, info.Name())
		info = file.NewInfo(filename, info.Size(), info.Mode(), info.ModTime(), info.IsDir(), options...)
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Typeflag = tar.TypeDir
		}
		if err = writer.WriteHeader(header); err != nil {
			return err
		}
		if info.Mode().IsRegular() && reader != nil {
			_, err = io.Copy(writer, reader)
		}
		return err

	}, writer, nil
}

//NewBatchUploader returns a batch uploader
func NewBatchUploader(dest storage.Uploader) storage.BatchUploader {
	return &uploader{uploader: dest}
}
