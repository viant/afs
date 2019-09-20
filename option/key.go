package option

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

//AES256Key represents custom key
type AES256Key struct {
	Key                 []byte
	Base64Key           string
	Base64KeyMd5Hash    string
	Base64KeySha256Hash string
}

//Init initialises key
func (k *AES256Key) Init() (err error) {
	if k.Base64Key != "" && len(k.Key) == 0 {
		if k.Key, err = base64.StdEncoding.DecodeString(k.Base64Key); err != nil {
			return err
		}
	} else if k.Base64Key == "" && len(k.Key) > 0 {
		k.Base64Key = base64.StdEncoding.EncodeToString(k.Key)
	}
	if k.Base64KeyMd5Hash == "" {
		md5keyHash := md5.New()
		md5keyHash.Write(k.Key)
		k.Base64KeyMd5Hash = base64.StdEncoding.EncodeToString(md5keyHash.Sum(nil))
	}
	if k.Base64KeySha256Hash == "" {
		sha256keyHash := sha256.Sum256(k.Key)
		k.Base64KeySha256Hash = base64.StdEncoding.EncodeToString(sha256keyHash[:])
	}
	return err
}

//Validate checks if key is valid
func (k *AES256Key) Validate() error {
	if len(k.Key) != 32 {
		return fmt.Errorf("%s: not a 32-byte AES-256 key", k.Key)
	}
	return nil
}

//NewAES256Key returns new key
func NewAES256Key(key []byte) (result *AES256Key, err error) {
	result = &AES256Key{Key: key}
	if err = result.Init(); err == nil {
		err = result.Validate()
	}
	return result, err
}

//NewBase64AES256Key create a AES256Key from base64 encoded key
func NewBase64AES256Key(base64Key string) (result *AES256Key, err error) {
	result = &AES256Key{Base64Key: base64Key}
	if err = result.Init(); err == nil {
		err = result.Validate()
	}
	return result, err
}
