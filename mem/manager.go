package mem

import (
	"context"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io"
	"net/http"
	"os"
	"strings"
)

type manager struct {
	*base.Manager
}

func (m *manager) provider(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	return NewStorager(baseURL), nil
}

func (m *manager) ErrorCode(err error) int {
	if err == nil {
		return 0
	}
	if strings.Contains(err.Error(), preconditionErrorMessage) {
		return http.StatusPreconditionFailed
	}
	if strings.Contains(err.Error(), noSuchFileOrDirectoryErrorMessage) {
		return http.StatusNotFound
	}

	return 0
}

func (m *manager) setErrors(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options []storage.Option) error {
	errors := option.Errors{}
	optError := &option.Error{}
	option.Assign(options, &errors, &optError)
	if optError.Type != "" && len(errors) == 0 {
		errors = append(errors, optError)
	}
	if len(errors) == 0 {
		return nil
	}
	if objects, err := m.List(ctx, URL); err == nil && len(objects) == 1 {
		file := &File{}
		if err = objects[0].Unwrap(&file); err != nil {
			return err
		}
		file.SetErrors(errors...)
		if file.uploadError != nil {
			return file.uploadError
		}
	}
	return nil
}

func (m *manager) Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	err := m.Manager.Upload(ctx, URL, mode, reader, options...)
	if err == nil {
		err = m.setErrors(ctx, URL, mode, reader, options)
	}
	return err
}

//New create a in memory storage
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}

func newManager(options ...storage.Option) *manager {
	result := &manager{}
	baseMgr := base.New(result, Scheme, result.provider, options)
	result.Manager = baseMgr
	return result
}
