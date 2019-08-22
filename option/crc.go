package option

import (
	"encoding/base64"
	"fmt"
	"hash/crc32"
)

//Crc represents crc hash
type Crc struct {
	Hash uint32
}

//Encode encodes hash
func (c *Crc) Encode() string {
	b := []byte{byte(c.Hash >> 24), byte(c.Hash >> 16), byte(c.Hash >> 8), byte(c.Hash)}
	return base64.StdEncoding.EncodeToString(b)
}

//Decode decodes base64 encoded hash
func (c *Crc) Decode(encoded string) error {
	d, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	if len(d) != 4 {
		return fmt.Errorf("storage: %q does not encode a 32-bit value", d)
	}
	c.Hash = uint32(d[0])<<24 + uint32(d[1])<<16 + uint32(d[2])<<8 + uint32(d[3])
	return nil
}

//NewCrc returns a crc hash for supplied data
func NewCrc(data []byte) *Crc {
	crc32Hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	_, _ = crc32Hash.Write(data)
	return &Crc{Hash: crc32Hash.Sum32()}
}
