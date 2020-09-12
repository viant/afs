package file

import (
	"context"
	"github.com/viant/afs/storage"
	"io"
	"os"
)

type manager struct{}

func (s *manager) List(ctx context.Context, URL string, options ...storage.Option) ([]storage.Object, error) {
	return List(ctx, URL, options...)
}

func (s *manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	return Upload(ctx, URL, mode, reader, options...)
}

func (s *manager) Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return Open(ctx, object, options...)
}

func (s *manager) OpenURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	return OpenURL(ctx, URL, options...)
}

func (s *manager) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	return Delete(ctx, URL, options...)
}

func (s *manager) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	return Create(ctx, URL, mode, isDir, options)
}

func (s *manager) Move(ctx context.Context, sourceURL, destURL string, options ...storage.Option) error {
	return Move(ctx, sourceURL, destURL, options...)
}

func (s *manager) NewWriter(_ context.Context, URL string, mode os.FileMode, options ...storage.Option) (io.WriteCloser, error) {
	return NewWriter(nil, URL, mode, options...)
}

func (s *manager) Close() error {
	return nil
}

func (s *manager) Scheme() string {
	return Scheme
}

//New returns a file manager
func New() storage.Manager {
	return &manager{}
}
