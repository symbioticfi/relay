package crypto

import (
	"github.com/go-errors/errors"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto/blsBn254"
	"middleware-offchain/core/usecase/crypto/ecdsaSecp256k1"
	"middleware-offchain/core/usecase/crypto/key-types"
)

type PublicKey = key_types.PublicKey
type PrivateKey = key_types.PrivateKey

func NewPublicKey(keyTag entity.KeyTag, keyBytes entity.RawPublicKey) (PublicKey, error) {
	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		return blsBn254.FromRaw(keyBytes)
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.FromRaw(keyBytes)
	case entity.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func NewPrivateKey(keyTag entity.KeyTag, keyBytes []byte) (PrivateKey, error) {
	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		return blsBn254.NewPrivateKey(keyBytes)
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.NewPrivateKey(keyBytes)
	case entity.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}
