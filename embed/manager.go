package embed

import (
	"context"
	"embed"
	"fmt"
	"github.com/viant/afs/storage"
	"io"
	"os"
)

type manager struct {
	fs  *embed.FS
	err error
}

func (s *manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	return fmt.Errorf("unsupproted Upload operation for %v", URL)
}

func (s *manager) Open(ctx context.Context, object storage.Object, options ...storage.Option) (io.ReadCloser, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.OpenURL(ctx, object.URL(), options...)
}

// Delete unsupported
func (s *manager) Delete(ctx context.Context, URL string, options ...storage.Option) error {
	return fmt.Errorf("unsupproted Delete operation for %v", URL)
}

// Create unsupported
func (s *manager) Create(ctx context.Context, URL string, mode os.FileMode, isDir bool, options ...storage.Option) error {
	return fmt.Errorf("unsupproted Create operation for %v", URL)
}

// Close closes mananger
func (s *manager) Close() error {
	return nil
}

// Scheme returns schmea
func (s *manager) Scheme() string {
	return Scheme
}

func newManager(options ...storage.Option) *manager {
	var fs *embed.FS

	for _, option := range options {
		switch v := option.(type) {
		case *embed.FS:
			fs = v
		case embed.FS:
			fs = &v
		}
	}
	var err error
	if fs == nil {
		err = fmt.Errorf("expcted %T", fs)
	}
	return &manager{
		fs:  fs,
		err: err,
	}
}

// New creates HTTP manager
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}
