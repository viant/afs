package option

//ServerSideEncryption represents server side encryption
type ServerSideEncryption struct {
	Algorithm string
}

//NewServerSideEncryption creates a server side encryption
func NewServerSideEncryption(alg string) *ServerSideEncryption {
	return &ServerSideEncryption{Algorithm: alg}
}
