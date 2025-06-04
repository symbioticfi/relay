package signer

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	keyprovider "middleware-offchain/pkg/key-provider"
)

type Signer struct {
	kp keyprovider.KeyProvider
}

func NewSigner(kp keyprovider.KeyProvider) *Signer {
	return &Signer{kp: kp}
}

func (s *Signer) Hash(keyTag uint8, message []byte) ([]byte, error) {
	keyType := keyTag >> 4
	if keyType == keyprovider.KeyTypeBlsBn254 {
		return crypto.Keccak256(message), nil
	}
	if keyType == keyprovider.KeyTypeEcdsaSecp256k1 {
		return crypto.Keccak256(message), nil
	}

	return nil, errors.New("invalid key type")
}

func (s *Signer) Sign(keyTag uint8, message []byte) (entity.Signature, error) {
	pk, err := s.kp.GetPrivateKey(keyTag)
	if err != nil {
		return entity.Signature{}, err
	}

	keyType := keyTag >> 4
	hash, err := s.Hash(keyTag, message)
	if err != nil {
		return entity.Signature{}, err
	}

	switch keyType {
	case keyprovider.KeyTypeBlsBn254:
		keyPair := bls.ComputeKeyPair(pk)
		blsSig, err := keyPair.Sign(hash)
		if err != nil {
			return entity.Signature{}, err
		}

		sig := entity.Signature{
			MessageHash: hash,
			Signature:   blsSig.Marshal(),
			PublicKey:   keyPair.PackPublicG1G2(),
		}

		return sig, nil

	case keyprovider.KeyTypeEcdsaSecp256k1:
		// same but for another key type
	}

	// assert, should not reach the code
	return entity.Signature{}, errors.Errorf("unsupported key type: %d", keyType)
}
