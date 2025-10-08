package crypto

import (
	"github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/ecdsaSecp256k1"
	key_types "github.com/symbioticfi/relay/symbiotic/usecase/crypto/key-types"

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

func HashMessage(keyType entity.KeyType, msg []byte) (entity.RawMessageHash, error) {
	switch keyType {
	case entity.KeyTypeBlsBn254:
		return blsBn254.HashMessage(msg), nil
	case entity.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.HashMessage(msg), nil
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
