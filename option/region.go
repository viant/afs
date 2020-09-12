package option

//Region represents cloud region/location option
type Region struct {
	Name string
}

//NewRegion creates a region for specified name
func NewRegion(name string) *Region {
	return &Region{Name: name}
}
