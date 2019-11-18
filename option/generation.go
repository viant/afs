package option

//Generation represent generation option
type Generation struct {
	WhenMatch  bool
	Generation int64
}

//NewGeneration create a generation
func NewGeneration(whenMatch bool, generation int64) *Generation {
	return &Generation{
		WhenMatch:  whenMatch,
		Generation: generation,
	}
}
