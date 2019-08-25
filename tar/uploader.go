package tar

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"os"
	"path"
)

type uploader struct {
	uploader storage.Uploader
	buffer   *bytes.Buffer
}

func (u *uploader) Uploader(ctx context.Context, URL string, options ...storage.Option) (storage.Upload, io.Closer, error) {
	var uploader storage.Uploader
	option.Assign(options, &u.buffer, &uploader)
	if uploader == nil {
		uploader = u.uploader
	}
	if u.buffer == nil {
		u.buffer = new(bytes.Buffer)
	}
	writer := newWriter(ctx, u.buffer, URL, uploader)
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

//newBatchUploader returns a batch uploader
func newBatchUploader(dest storage.Uploader) *uploader {
	return &uploader{uploader: dest}
}

//NewBatchUploader returns a batch uploader
func NewBatchUploader(dest storage.Uploader) storage.BatchUploader {
	return newBatchUploader(dest)
}
