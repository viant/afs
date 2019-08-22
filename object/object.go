package object

import (
	"fmt"
	"github.com/viant/afs/storage"
	"os"
	"reflect"
)

//Object represents abstract storage object
type Object struct {
	url      string
	Source   interface{}
	linkname string
	linkURL  string
	os.FileInfo
}

//URL return storage url
func (o *Object) URL() string {
	return o.url
}

//Linkname returns a link name
func (o *Object) Linkname() string {
	return o.linkname
}

//LinkURL returns link URL (absolute path)
func (o *Object) LinkURL() string {
	return o.linkURL
}

//Wrap wraps Source storage object
func (o *Object) Wrap(source interface{}) {
	o.Source = source
}

//Unwrap unwrap source storage to target pointer
func (o *Object) Unwrap(target interface{}) error {
	if o.Source == nil {
		return nil
	}
	targetValue := reflect.ValueOf(target)
	sourceValue := reflect.ValueOf(o.Source)

	if sourceValue.Type().AssignableTo(targetValue.Type()) {
		return fmt.Errorf("unable to assign %T to %T", o.Source, target)
	}
	targetValue.Elem().Set(sourceValue)
	return nil
}

//New creates a new storage object
func New(URL string, info os.FileInfo, source interface{}) storage.Object {
	linkname := ""
	linkURL := ""
	link, ok := source.(*Link)
	if ok {
		linkname = link.Linkname
		linkURL = link.LinkURL
		source = link.Source
	}
	var result = &Object{
		url:      URL,
		Source:   source,
		linkname: linkname,
		linkURL:  linkURL,
		FileInfo: info,
	}
	return result
}
