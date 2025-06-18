package entity

import "fmt"

type KeyType uint8

const (
	KeyTypeBlsBn254       KeyType = 0
	KeyTypeEcdsaSecp256k1 KeyType = 1

	KeyTypeInvalid KeyType = 255
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
	return nil, fmt.Errorf("unsupported key type: %d", keyType)
}
