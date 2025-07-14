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

func NewPublicKey(keyType entity.KeyType, keyBytes entity.RawPublicKey) (PublicKey, error) {
	switch keyType {
	case entity.KeyTypeBlsBn254:
		return blsBn254.FromRaw(keyBytes)
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.FromRaw(keyBytes)
	case entity.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func NewPrivateKey(keyType entity.KeyType, keyBytes []byte) (PrivateKey, error) {
	switch keyType {
	case entity.KeyTypeBlsBn254:
		return blsBn254.NewPrivateKey(keyBytes)
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.NewPrivateKey(keyBytes)
	case entity.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func GeneratePrivateKey(keyType entity.KeyType) (PrivateKey, error) {
	switch keyType {
	case entity.KeyTypeBlsBn254:
		return blsBn254.GenerateKey()
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.GenerateKey()
	case entity.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}
