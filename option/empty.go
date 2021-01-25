package option

//Empty represents empty pipeline writer option
type Empty struct {
	Allowed bool
}

//NewEmpty creates a new empty option
func NewEmpty(allowed bool) *Empty {
	return &Empty{Allowed: allowed}
}
