package option

//SkipChecksum represents checksum option
type SkipChecksum struct {
	Skip bool
}

//NewSkipChecksum returns checksum options for supplied skip flag
func NewSkipChecksum(skip bool) *SkipChecksum {
	return &SkipChecksum{Skip: skip}
}
