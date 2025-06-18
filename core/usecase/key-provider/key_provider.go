package keyprovider

import (
	"errors"
	"strconv"

	"middleware-offchain/core/entity"
)

func typeToStr(keyType entity.KeyType) (string, error) {
	switch keyType {
	case entity.KeyTypeBlsBn254:
		return "BLS-BN254", nil
	case entity.KeyTypeEcdsaSecp256k1:
		return "ECDSA-SECP256K1", nil
	case entity.KeyTypeInvalid:
		return "Invalid", nil
	}
	return "", errors.New("invalid key type")
}

func getAlias(keyTag entity.KeyTag) (string, error) {
	// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/contracts/libraries/utils/KeyTags.sol#L24-L40
	keyId := keyTag & 0x0F

	keyTypeStr, err := typeToStr(keyTag.Type())
	if err != nil {
		return "", err
	}

	ketIdStr := strconv.Itoa(int(keyId))

	return "KEY-SYMB-" + keyTypeStr + "-" + ketIdStr, nil
}
