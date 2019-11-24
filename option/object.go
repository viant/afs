package option

//ObjectKind represents an option to indicate operation object kind
type ObjectKind struct {
	File bool
}

//NewObject creates a new object
func NewObjectKind(file bool) *ObjectKind {
	return &ObjectKind{File: file}
}
