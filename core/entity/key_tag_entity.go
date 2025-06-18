package entity

import "errors"

type KeyType uint8

const (
	KeyTypeBlsBn254       KeyType = 0
	KeyTypeEcdsaSecp256k1 KeyType = 1

	BLS_BN254_TYPE       = "bls_bn254"
	ECDSA_SECP256K1_TYPE = "ecdsa_secp256k1"
)

type KeyTag uint8

func (kt KeyTag) Type() KeyType {
	switch uint8(kt) >> 4 {
	case 0:
		return KeyTypeBlsBn254
	case 1:
		return KeyTypeEcdsaSecp256k1
	default:
		return 0 // Invalid key type
	}
}

func (kt KeyType) String() (string, error) {
	switch kt {
	case KeyTypeBlsBn254:
		return BLS_BN254_TYPE, nil
	case KeyTypeEcdsaSecp256k1:
		return ECDSA_SECP256K1_TYPE, nil
	}
	return "", errors.New("invalid key type")
}

func KeyTypeFromString(typeStr string) (KeyType, error) {
	switch typeStr {
	case BLS_BN254_TYPE:
		return KeyTypeBlsBn254, nil
	case ECDSA_SECP256K1_TYPE:
		return KeyTypeEcdsaSecp256k1, nil
	}
	return 0, errors.New("invalid key type")
}
