package signer

import (
	"errors"

	"github.com/ethereum/go-ethereum/crypto"

	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/key-provider"
	"middleware-offchain/pkg/types"
)

type Signer struct {
	Kp keyprovider.KeyProvider
}

func NewSigner(kp keyprovider.KeyProvider) *Signer {
	return &Signer{Kp: kp}
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

func (s *Signer) Sign(keyTag uint8, message []byte) (*types.Signature, error) {
	pk, err := s.Kp.GetPrivateKey(keyTag)
	if err != nil {
		return nil, err
	}

	keyType := keyTag >> 4
	hash, err := s.Hash(keyTag, message)
	if err != nil {
		return nil, err
	}

	if keyType == keyprovider.KeyTypeBlsBn254 {
		keyPair := bls.ComputeKeyPair(pk)
		blsSig, err := keyPair.Sign(hash)
		if err != nil {
			return nil, err
		}

		sig := &types.Signature{}
		sig.MessageHash = hash
		sig.Signature = blsSig.Marshal()
		sig.PublicKey = keyPair.PackPublicG1G2()
		return sig, nil
	}
	if keyType == keyprovider.KeyTypeEcdsaSecp256k1 {
		// same but for another key type
	}

	// assert, should not reach the code
	return nil, errors.New("invalid key type")
}
