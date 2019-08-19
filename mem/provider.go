package mem

import (
	"github.com/viant/afs/storage"
)

func Provider(options ...storage.Option) (storage.Manager, error) {
	return Singleton(options...), nil
}
