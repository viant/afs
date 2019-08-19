package file

import (
	"github.com/viant/afs/storage"
)

func Provider(options ...storage.Option) (storage.Manager, error) {
	return New(), nil
}
