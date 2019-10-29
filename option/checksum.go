package option

//Checksum represents checksum option
type Checksum struct {
	Skip bool
}

//NewChecksum returns checksum options for supplied skip flag
func NewChecksum(skip bool) *Checksum{
	return &Checksum{Skip:skip}
}