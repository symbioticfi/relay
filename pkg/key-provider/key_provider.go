package keyprovider

import (
	"errors"
	"strconv"
)

const (
	KeyTypeBlsBn254       uint8 = 0
	KeyTypeEcdsaSecp256k1 uint8 = 1
)

func typeToStr(keyType uint8) (string, error) {
	switch keyType {
	case KeyTypeBlsBn254:
		return "BLS-BN254", nil
	case KeyTypeEcdsaSecp256k1:
		return "ECDSA-SECP256K1", nil
	}
	return "", errors.New("invalid key type")
}

func getAlias(keyTag uint8) (string, error) {
	// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/contracts/libraries/utils/KeyTags.sol#L24-L40
	keyType := keyTag >> 4
	keyId := keyTag & 0x0F

	keyTypeStr, err := typeToStr(keyType)
	if err != nil {
		return "", err
	}

	ketIdStr := strconv.Itoa(int(keyId))

	return "KEY-SYMB-" + keyTypeStr + "-" + ketIdStr, nil
}
