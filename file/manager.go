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

func (s *manager) Download(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	return Download(ctx, object, options...)
}

func (s *manager) DownloadWithURL(ctx context.Context, URL string, options ...storage.Option) (io.ReadCloser, error) {
	return DownloadWithURL(ctx, URL, options...)
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

func (s *manager) Close() error {
	return nil
}

func (s *manager) Scheme() string {
	return Scheme
}

func New() storage.Manager {
	return &manager{}
}
