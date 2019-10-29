package option

//Stream represents stream option for download reader
type Stream struct {
	PartSize int
	Size     int
}

//NewStream returns a new stream
func NewStream(partSize, size int) *Stream {
	return &Stream{
		PartSize: partSize,
		Size:     size,
	}
}
