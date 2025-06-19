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

//
//func (s *Signer) Hash(keyTag entity.KeyTag, message []byte) ([]byte, error) {
//	switch keyTag.Type() {
//	case entity.KeyTypeBlsBn254:
//		return crypto.Keccak256(message), nil
//	case entity.KeyTypeEcdsaSecp256k1:
//		return crypto.Keccak256(message), nil
//	case entity.KeyTypeInvalid:
//		return nil, errors.New("invalid key type")
//	}
//	return nil, errors.New("invalid key type")
//}
//
//// Verify returns the compressed public key, for bls it will be G1 point
//func (s *Signer) Verify(keyTag entity.KeyTag, signature entity.SignatureExtended) ([]byte, bool, error) {
//	switch keyTag.Type() {
//	case entity.KeyTypeBlsBn254:
//		g1PubKey, g2PubKey, err := bls.UnpackPublicG1G2(signature.PublicKey)
//		if err != nil {
//			return nil, false, err
//		}
//
//		g1Sig, err := bls.DeserializeG1(signature.Signature)
//		if err != nil {
//			return nil, false, err
//		}
//
//		ok, err := bls.Verify(&g2PubKey, g1Sig, signature.MessageHash)
//		if err != nil {
//			return nil, false, err
//		}
//
//		return g1PubKey.Marshal(), ok, nil
//	case entity.KeyTypeEcdsaSecp256k1:
//		return nil, false, nil
//	case entity.KeyTypeInvalid:
//		return nil, false, errors.Errorf("unsupported key type: %d", keyTag.Type())
//	}
//	return nil, false, errors.Errorf("unsupported key type: %d", keyTag.Type())
//}
//
//func (s *Signer) Sign(keyTag entity.KeyTag, message []byte) (entity.SignatureExtended, error) {
//	pk, err := s.kp.GetPrivateKey(keyTag)
//	if err != nil {
//		return entity.SignatureExtended{}, err
//	}
//
//	hash, err := s.Hash(keyTag, message)
//	if err != nil {
//		return entity.SignatureExtended{}, err
//	}
//
//	switch keyTag.Type() {
//	case entity.KeyTypeBlsBn254:
//		keyPair := bls.ComputeKeyPair(pk)
//		blsSig, err := keyPair.Sign(hash)
//		if err != nil {
//			return entity.SignatureExtended{}, err
//		}
//
//		sig := entity.SignatureExtended{
//			MessageHash: hash,
//			Signature:   blsSig.Marshal(),
//			PublicKey:   keyPair.PackPublicG1G2(),
//		}
//
//		return sig, nil
//
//	case entity.KeyTypeEcdsaSecp256k1:
//		// same but for another key type
//	case entity.KeyTypeInvalid:
//		return entity.SignatureExtended{}, errors.Errorf("unsupported key type: %d", keyTag.Type())
//	}
//
//	// assert, should not reach the code
//	return entity.SignatureExtended{}, errors.Errorf("unsupported key type: %d", keyTag.Type())
//}
//
//func (s *Signer) GetPublicKey(keyTag entity.KeyTag) ([]byte, error) {
//	sk, err := s.kp.GetPrivateKey(keyTag)
//	if err != nil {
//		return nil, err
//	}
//
//	switch keyTag.Type() {
//	case entity.KeyTypeBlsBn254:
//		kp := bls.ComputeKeyPair(sk)
//		return kp.PublicKeyG1.Marshal(), nil
//	case entity.KeyTypeEcdsaSecp256k1:
//		return nil, errors.New("ECDSA key type not supported in this provider")
//	case entity.KeyTypeInvalid:
//		return sk, nil
//	}
//
//	return sk, nil
//}
