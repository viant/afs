package option

//Location represents a location
type Location struct {
	Path string
}

//NewLocation create a location with supplied path
func NewLocation(path string) *Location {
	return &Location{Path: path}
}
