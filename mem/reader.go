package mem

type fakeReader struct {
	error error
}

//Read returns defined error
func (r *fakeReader) Read(p []byte) (n int, err error) {
	return 0, r.error
}
