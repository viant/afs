package tar

import "github.com/viant/afs/storage"

//Provider returns a http manager
func Provider(options ...storage.Option) (storage.Manager, error) {
	return New(options...), nil
}
