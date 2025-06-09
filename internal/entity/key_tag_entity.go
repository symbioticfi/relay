package entity

type KeyType uint8

const (
	KeyTypeBlsBn254       KeyType = 0
	KeyTypeEcdsaSecp256k1 KeyType = 1
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
