package entity

import (
	"fmt"

	"github.com/go-errors/errors"
)

type KeyType uint8

const (
	KeyTypeBlsBn254       KeyType = 0
	KeyTypeEcdsaSecp256k1 KeyType = 1
	KeyTypeBls12381Bn254  KeyType = 2
	KeyTypeInvalid        KeyType = 255

	BLS_BN254_TYPE       = "bls_bn254"
	ECDSA_SECP256K1_TYPE = "ecdsa_secp256k1"
	BLS_12381_BN254_TYPE = "bls12381_bn254"
	INVALID_TYPE         = "invalid"
)

type KeyTag uint8

func (kt KeyTag) Type() KeyType {
	switch uint8(kt) >> 4 {
	case 0:
		return KeyTypeBlsBn254
	case 1:
		return KeyTypeEcdsaSecp256k1
	case 2:
		return KeyTypeBls12381Bn254
	default:
		return KeyTypeInvalid // Invalid key type
	}
}

func (kt KeyTag) MarshalText() (text []byte, err error) {
	keyType := kt.Type()
	keyID := uint8(kt) & 0x0F
	switch keyType {
	case KeyTypeBlsBn254:
		return []byte(fmt.Sprintf("%d (BLS-BN254/%d)", uint8(kt), keyID)), nil
	case KeyTypeEcdsaSecp256k1:
		return []byte(fmt.Sprintf("%d (ECDSA-SECP256K1/%d)", uint8(kt), keyID)), nil
	case KeyTypeBls12381Bn254:
		return []byte(fmt.Sprintf("%d (BLS12381-BN254/%d)", uint8(kt), keyID)), nil
	case KeyTypeInvalid:
		return []byte(fmt.Sprintf("%d (UNKNOWN/%d)", uint8(kt), keyID)), nil
	}
	return nil, errors.Errorf("unsupported key type: %d", keyType)
}

func (kt KeyTag) String() string {
	keyType := kt.Type()
	keyTag := uint8(kt) & 0x0F
	switch keyType {
	case KeyTypeBlsBn254:
		return fmt.Sprintf("%d (BLS-BN254/%d)", uint8(kt), keyTag)
	case KeyTypeEcdsaSecp256k1:
		return fmt.Sprintf("%d (ECDSA-SECP256K1/%d)", uint8(kt), keyTag)
	case KeyTypeBls12381Bn254:
		return fmt.Sprintf("%d (BLS12381-BN254/%d)", uint8(kt), keyTag)
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
	case KeyTypeBls12381Bn254:
		return BLS_12381_BN254_TYPE, nil
	case KeyTypeInvalid:
		return INVALID_TYPE, nil
	}
	return "", errors.New("invalid key type")
}

// SignerKey returns true if the key type can be used for signing
func (kt KeyType) SignerKey() bool {
	switch kt {
	case KeyTypeBlsBn254, KeyTypeEcdsaSecp256k1, KeyTypeBls12381Bn254:
		return true
	case KeyTypeInvalid:
		return false
	}
	return false
}

// AggregationKey returns true if the key type can be used for aggregation
func (kt KeyType) AggregationKey() bool {
	switch kt {
	case KeyTypeBlsBn254:
		return true
	case KeyTypeEcdsaSecp256k1, KeyTypeBls12381Bn254, KeyTypeInvalid:
		return false
	}
	return false
}

func KeyTypeFromString(typeStr string) (KeyType, error) {
	switch typeStr {
	case BLS_BN254_TYPE:
		return KeyTypeBlsBn254, nil
	case ECDSA_SECP256K1_TYPE:
		return KeyTypeEcdsaSecp256k1, nil
	case BLS_12381_BN254_TYPE:
		return KeyTypeBls12381Bn254, nil
	case INVALID_TYPE:
		return KeyTypeInvalid, nil
	}
	return 0, errors.New("invalid key type")
}

func KeyTagFromTypeAndId(keyType KeyType, keyId uint8) (KeyTag, error) {
	if keyId > 15 {
		return 0, errors.New("key id must be between 0 and 15")
	}

	switch keyType {
	case KeyTypeBlsBn254:
	case KeyTypeEcdsaSecp256k1:
	case KeyTypeBls12381Bn254:
	case KeyTypeInvalid:
		return 0, errors.New("invalid key type")
	default:
		return 0, errors.New("unsupported key type")
	}

	return KeyTag((uint8(keyType) << 4) | keyId), nil
}
