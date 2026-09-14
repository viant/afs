package resource

import (
	"context"
	"io"
)

// Cancellation is checked between I/O calls. Blocking calls remain the backing
// implementation's responsibility; fs.FS has no context-aware I/O contract.
type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

type contextReadCloser struct {
	io.Reader
	io.Closer
}
type contextWriter struct {
	ctx context.Context
	io.WriteCloser
}

func (w *contextWriter) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	return w.WriteCloser.Write(p)
}
func (w *contextWriter) Close() error {
	err := w.WriteCloser.Close()
	if err != nil {
		return err
	}
	return w.ctx.Err()
}
