package option

import "github.com/viant/afs/storage"

//Dest represents dest options
type Dest storage.Options

//NewDest returns new source options
func NewDest(options ...storage.Option) *Dest {
	result := Dest(options)
	return &result
}
