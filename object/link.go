package object

//Link represents a link source wrapper
type Link struct {
	Source   interface{}
	Linkname string
	LinkURL  string
}

//NewLink create a link
func NewLink(linkname, linkURL string, source interface{}) *Link {
	return &Link{Linkname: linkname, LinkURL: linkURL, Source: source}
}
