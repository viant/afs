package scp

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/viant/afs/option"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

//AuthProvider represents ssh client config authProvider
type AuthProvider interface {
	ClientConfig() (*ssh.ClientConfig, error)
}

type authProvider struct {
	pemAuth  KeyAuth
	credAuth option.BasicAuth
}

//ClientConfig returns client config
func (p *authProvider) ClientConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            os.Getenv("USER"),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            make([]ssh.AuthMethod, 0),
	}
	if p.pemAuth != nil {
		config.User = p.pemAuth.Username()
		key, err := p.pemAuth.Singer()
		if err != nil {
			return nil, err
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(key))
	}
	if p.credAuth != nil {
		user, password := p.credAuth.Credentials()
		config.User = user
		config.Auth = append(config.Auth, ssh.Password(password))
	}
	return config, nil
}

//NewAuthProvider returns new auth provider
func NewAuthProvider(pemAuth KeyAuth, credAuth option.BasicAuth) AuthProvider {
	return &authProvider{pemAuth: pemAuth, credAuth: credAuth}
}

//KeyAuth represents a key based auth
type KeyAuth interface {
	//Singer returns signer key
	Singer() (ssh.Signer, error)

	Username() string
}

type keyAuthnticator struct {
	keyLocation  string
	username     string
	_keyPassword string
}

//Username returns a username
func (a *keyAuthnticator) Username() string {
	return a.username
}

func (a *keyAuthnticator) isKeyEncrypted(data []byte) bool {
	block, _ := pem.Decode(data)
	if block == nil {
		return false
	}
	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED")
}

func (a *keyAuthnticator) Singer() (ssh.Signer, error) {
	rawPEM, err := a.pem()
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(rawPEM)
}

//PEM returns secure key container
func (a *keyAuthnticator) pem() ([]byte, error) {
	pemBytes, err := ioutil.ReadFile(a.keyLocation)
	if err != nil {
		return nil, err
	}

	if a.isKeyEncrypted(pemBytes) {
		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return nil, fmt.Errorf("unable decode %v", a.keyLocation)
		}
		if x509.IsEncryptedPEMBlock(block) {
			key, err := x509.DecryptPEMBlock(block, []byte(a._keyPassword))
			if err != nil {
				return nil, err
			}
			block = &pem.Block{Type: block.Type, Bytes: key}
			pemBytes = pem.EncodeToMemory(block)
			return pemBytes, nil
		}
	}
	return pemBytes, nil
}

//NewKeyAuth returns a new private key authenticator
func NewKeyAuth(keyLocation, username, keyPassword string) KeyAuth {
	return &keyAuthnticator{
		keyLocation:  keyLocation,
		username:     username,
		_keyPassword: keyPassword,
	}
}

//LocalhostKeyAuth returns a localhost key authenticator with ~/.ssh/authorized_keys
func LocalhostKeyAuth(keyPassword string, locations ...string) (KeyAuth, error) {
	username := os.Getenv("USER")
	if username == "" {
		return nil, fmt.Errorf("username was empty")
	}
	locations = append(locations, path.Join(os.Getenv("HOME"), ".secret", "id_rsa"))
	locations = append(locations, path.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
	keyLocation := ""
	for _, candidate := range locations {
		if _, err := os.Stat(candidate); err == nil {
			keyLocation = candidate
			break
		}
	}
	if keyLocation == "" {
		return nil, fmt.Errorf("failed to lookup key location: %v", locations)
	}
	return NewKeyAuth(keyLocation, username, keyPassword), nil
}
