package option

import "github.com/viant/afs/storage"

//Append storage options
func Append(options []storage.Option, newOptions ...storage.Option) []storage.Option {
	if len(options) == 0 {
		return newOptions
	}
	return append(options, newOptions...)
}
