package option

//Grant represents a grant option
type Grant struct {
	FullControl string
	Read        string
	ReadACP     string
	WriteACP    string
}

//NewGrant creates a grant option
func NewGrant(fullControl, read, readACP, writeACP string) *Grant {
	return &Grant{
		FullControl: fullControl,
		Read: read,
		ReadACP: readACP,
		WriteACP: writeACP,
	}
}
