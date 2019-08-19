package scp

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func newTestClient(address string) (*ssh.Client, error) {
	authenticator, err := LocalhostKeyAuth("")
	if err != nil {
		return nil, err
	}
	provider := NewAuthProvider(authenticator, nil)
	config, err := provider.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create *ssh.ClientConfig")
	}
	return ssh.Dial("tcp", address, config)
}
