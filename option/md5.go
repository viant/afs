package option

import (
	"crypto/md5"
	"encoding/base64"
)

//Md5 represents md5 value
type Md5 struct {
	Hash []byte
}

//Encode encode base64 hash value
func (m *Md5) Encode() string {
	return base64.StdEncoding.EncodeToString(m.Hash)
}

//Decode base64 decode
func (m *Md5) Decode(encoded string) (err error) {
	m.Hash, err = base64.StdEncoding.DecodeString(encoded)
	return err
}

//NewMd5 returns a MD5 hash for supplied data
func NewMd5(data []byte) *Md5 {
	hash := md5.New()
	_, _ = hash.Write(data)
	return &Md5{Hash: hash.Sum(nil)}
}
