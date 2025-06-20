package keyprovider

import (
	"errors"
	"strconv"
	"strings"

	"middleware-offchain/core/entity"
)

func getAlias(keyTag entity.KeyTag) (string, error) {
	// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/contracts/libraries/utils/KeyTags.sol#L24-L40
	keyId := keyTag & 0x0F

	keyTypeStr, err := keyTag.Type().String()
	if err != nil {
		return "", err
	}

	keyIdStr := strconv.Itoa(int(keyId))

	return "key-symb-" + keyTypeStr + "-" + keyIdStr, nil
}

func AliasToTag(alias string) (entity.KeyTag, error) {
	keyTagParts := strings.Split(alias, "-")
	if len(keyTagParts) != 4 {
		return 0, errors.New("invalid alias")
	}

	keyType, err := entity.KeyTypeFromString(keyTagParts[2])
	if err != nil {
		return 0, err
	}

	keyId, err := strconv.Atoi(keyTagParts[3])
	if err != nil {
		return 0, err
	}

	return entity.KeyTag(uint8(keyType)<<4 + uint8(keyId)), nil
}
