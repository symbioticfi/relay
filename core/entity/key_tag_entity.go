package entity

import (
	"fmt"

	"github.com/go-errors/errors"
)

type KeyType uint8

const (
	KeyTypeBlsBn254       KeyType = 0
	KeyTypeEcdsaSecp256k1 KeyType = 1
	KeyTypeInvalid        KeyType = 255

	BLS_BN254_TYPE       = "bls_bn254"
	ECDSA_SECP256K1_TYPE = "ecdsa_secp256k1"
	INVALID_TYPE         = "invalid"
)

type KeyTag uint8

func (kt KeyTag) Type() KeyType {
	switch uint8(kt) >> 4 {
	case 0:
		return KeyTypeBlsBn254
	case 1:
		return KeyTypeEcdsaSecp256k1
	default:
		return KeyTypeInvalid // Invalid key type
	}
}

func (kt KeyTag) MarshalText() (text []byte, err error) {
	keyType := kt.Type()
	keyTag := uint8(keyType) & 0x0F
	switch keyType {
	case KeyTypeBlsBn254:
		return []byte(fmt.Sprintf("%d (BLS-BN254/%d)", uint8(kt), keyTag)), nil
	case KeyTypeEcdsaSecp256k1:
		return []byte(fmt.Sprintf("%d (ECDSA-SECP256K1/%d)", uint8(kt), keyTag)), nil
	case KeyTypeInvalid:
		return []byte(fmt.Sprintf("%d (UNKNOWN/%d)", uint8(kt), keyTag)), nil
	}
	return nil, errors.Errorf("unsupported key type: %d", keyType)
}

func (kt KeyTag) String() string {
	keyType := kt.Type()
	keyTag := uint8(keyType) & 0x0F
	switch keyType {
	case KeyTypeBlsBn254:
		return fmt.Sprintf("%d (BLS-BN254/%d)", uint8(kt), keyTag)
	case KeyTypeEcdsaSecp256k1:
		return fmt.Sprintf("%d (ECDSA-SECP256K1/%d)", uint8(kt), keyTag)
	case KeyTypeInvalid:
		return fmt.Sprintf("%d (UNKNOWN/%d)", uint8(kt), keyTag)
	}
	return fmt.Sprintf("%d (UNKNOWN/%d)", uint8(kt), keyTag)
}

func (kt KeyType) String() (string, error) {
	switch kt {
	case KeyTypeBlsBn254:
		return BLS_BN254_TYPE, nil
	case KeyTypeEcdsaSecp256k1:
		return ECDSA_SECP256K1_TYPE, nil
	case KeyTypeInvalid:
		return INVALID_TYPE, nil
	}
	return "", errors.New("invalid key type")
}

func KeyTypeFromString(typeStr string) (KeyType, error) {
	switch typeStr {
	case BLS_BN254_TYPE:
		return KeyTypeBlsBn254, nil
	case ECDSA_SECP256K1_TYPE:
		return KeyTypeEcdsaSecp256k1, nil
	case INVALID_TYPE:
		return KeyTypeInvalid, nil
	}
	return 0, errors.New("invalid key type")
}
