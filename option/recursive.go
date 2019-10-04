package option

//Recursive represents recursive option
type Recursive struct {
	Flag bool
}

//NewRecursive returns a recursive option
func NewRecursive(flag bool) *Recursive {
	return &Recursive{
		Flag: flag,
	}
}
