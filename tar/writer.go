package tar

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/viant/afs/storage"
)

type writer struct {
	ctx    context.Context
	buffer *bytes.Buffer
	*tar.Writer
	destURL  string
	uploader storage.Uploader
}

func newWriter(ctx context.Context, buffer *bytes.Buffer, URL string, uploader storage.Uploader) *writer {
	if buffer == nil {
		buffer = new(bytes.Buffer)
	}
	return &writer{
		ctx:      ctx,
		buffer:   buffer,
		Writer:   tar.NewWriter(buffer),
		destURL:  URL,
		uploader: uploader,
	}
}

func (w *writer) Close() error {
	err := w.Writer.Flush()
	if err == nil {
		if err = w.Writer.Close(); err == nil {
			if w.uploader != nil {
				err = w.uploader.Upload(w.ctx, w.destURL, 0644, bytes.NewReader(w.buffer.Bytes()))
			}
		}
	}
	return err
}
