package storage

//Sizer represents abstraction returing a size
type Sizer interface {
	Size() int64
}
