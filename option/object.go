package option


//Object represents expect file on Object operation
type Object struct {
	File bool
}

//NewObject creates a new object
func NewObject(file bool) *Object {
	return &Object{File:file}
}