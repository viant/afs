package parrot

import (
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
)

//Pkg returns package name for location
func Pkg(location string) string {
	parent, _ := url.Split(location, file.Scheme)
	_, pkg := url.Split(parent, file.Scheme)
	return pkg
}
