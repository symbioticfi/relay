package crypto

import (
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/ecdsaSecp256k1"
	key_types "github.com/symbioticfi/relay/symbiotic/usecase/crypto/key-types"

	"github.com/go-errors/errors"
)

type PublicKey = key_types.PublicKey
type PrivateKey = key_types.PrivateKey

func NewPublicKey(keyType symbiotic.KeyType, keyBytes symbiotic.RawPublicKey) (PublicKey, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.FromRaw(keyBytes)
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.FromRaw(keyBytes)
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func NewPrivateKey(keyType symbiotic.KeyType, keyBytes []byte) (PrivateKey, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.NewPrivateKey(keyBytes)
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.NewPrivateKey(keyBytes)
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func HashMessage(keyType symbiotic.KeyType, msg []byte) (symbiotic.RawMessageHash, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.HashMessage(msg), nil
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.HashMessage(msg), nil
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func GeneratePrivateKey(keyType symbiotic.KeyType) (PrivateKey, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.GenerateKey()
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.GenerateKey()
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}
