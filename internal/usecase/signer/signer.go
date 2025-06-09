package signer

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
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

func (s *Signer) Hash(keyTag entity.KeyTag, message []byte) ([]byte, error) {
	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		return crypto.Keccak256(message), nil
	case entity.KeyTypeEcdsaSecp256k1:
		return crypto.Keccak256(message), nil
	}

	return nil, errors.New("invalid key type")
}

func (s *Signer) Verify(keyTag entity.KeyTag, signature entity.Signature) (bool, error) {
	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		_, g2PubKey, err := bls.UnpackPublicG1G2(signature.PublicKey)
		if err != nil {
			return false, err
		}
		g1Sig, err := bls.DeserializeG1(signature.Signature)
		if err != nil {
			return false, err
		}
		ok, err := bls.Verify(&g2PubKey, g1Sig, signature.MessageHash)
		if err != nil {
			return false, err
		}
		return ok, nil
	case entity.KeyTypeEcdsaSecp256k1:
		return true, nil
	}

	return false, errors.Errorf("unsupported key type: %d", keyTag.Type())
}

func (s *Signer) Sign(keyTag entity.KeyTag, message []byte) (entity.Signature, error) {
	pk, err := s.kp.GetPrivateKey(keyTag)
	if err != nil {
		return entity.Signature{}, err
	}

	hash, err := s.Hash(keyTag, message)
	if err != nil {
		return entity.Signature{}, err
	}

	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
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

	case entity.KeyTypeEcdsaSecp256k1:
		// same but for another key type
	}

	// assert, should not reach the code
	return entity.Signature{}, errors.Errorf("unsupported key type: %d", keyTag.Type())
}

func (s *Signer) GetPublic(keyTag entity.KeyTag) ([]byte, error) {
	sk, err := s.kp.GetPrivateKey(keyTag)
	if err != nil {
		return nil, err
	}

	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		kp := bls.ComputeKeyPair(sk)
		return kp.PackPublicG1G2(), nil
	case entity.KeyTypeEcdsaSecp256k1:
		return nil, errors.New("ECDSA key type not supported in this provider")
	}

	return sk, nil
}
