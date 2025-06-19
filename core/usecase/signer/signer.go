package signer

import (
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"
)

type keyProvider interface {
	GetPrivateKey(keyTag entity.KeyTag) ([]byte, error)
	HasKey(keyTag entity.KeyTag) (bool, error)
}

type Signer struct {
	kp keyProvider
}

func NewSigner(kp keyProvider) *Signer {
	return &Signer{kp: kp}
}

func (s *Signer) GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error) {
	keyBytes, err := s.kp.GetPrivateKey(keyTag)
	if err != nil {
		return nil, errors.Errorf("failed to get the private key for tag %d: %s", keyTag, err)
	}
	return crypto.NewPrivateKey(keyTag, keyBytes)
}
