package option

import "github.com/viant/afs/storage"

//Source represents source options
type Source storage.Options

//NewSource returns new source options
func NewSource(options ...storage.Option) *Source {
	result := Source(options)
	return &result
}
