package crypto

import (
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto/blsBn254"
	"middleware-offchain/core/usecase/crypto/ecdsaSecp256k1"
	key_types "middleware-offchain/core/usecase/crypto/key-types"

	"github.com/go-errors/errors"
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
