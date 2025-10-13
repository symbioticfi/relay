package keyprovider

import (
	"strconv"
	"strings"

	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

const (
	SYMBIOTIC_KEY_NAMESPACE = "symb"
	EVM_KEY_NAMESPACE       = "evm"
	P2P_KEY_NAMESPACE       = "p2p"

	// DEFAULT_EVM_CHAIN_ID chain id used to identify the default key for all chains
	DEFAULT_EVM_CHAIN_ID = 0

	P2P_HOST_IDENTITY_KEY_ID = 1
)

type KeyProvider interface {
	GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error)
	GetPrivateKeyByAlias(alias string) (crypto.PrivateKey, error)
	GetPrivateKeyByNamespaceTypeId(namespace string, keyType symbiotic.KeyType, id int) (crypto.PrivateKey, error)
	HasKey(keyTag symbiotic.KeyTag) (bool, error)
	HasKeyByAlias(alias string) (bool, error)
	HasKeyByNamespaceTypeId(namespace string, keyType symbiotic.KeyType, id int) (bool, error)
}

func KeyTagToAliasWithNS(namespace string, keyTag symbiotic.KeyTag) (string, error) {
	// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/contracts/libraries/utils/KeyTags.sol#L24-L40
	keyId := keyTag & 0x0F

	return ToAlias(namespace, keyTag.Type(), int(keyId))
}

func KeyTagToAlias(keyTag symbiotic.KeyTag) (string, error) {
	return KeyTagToAliasWithNS(SYMBIOTIC_KEY_NAMESPACE, keyTag)
}

func ToAlias(namespace string, keyType symbiotic.KeyType, keyId int) (string, error) {
	keyTypeStr, err := keyType.String()
	if err != nil {
		return "", err
	}

	if strings.Contains(namespace, "-") {
		return "", errors.New("namespace must not contain dash")
	}

	keyIdStr := strconv.Itoa(keyId)

	return namespace + "-" + keyTypeStr + "-" + keyIdStr, nil
}

//nolint:revive // we need to return the ns too
func AliasToKeyTypeId(alias string) (string, symbiotic.KeyType, int, error) {
	keyTagParts := strings.Split(alias, "-")
	if len(keyTagParts) != 3 {
		return "", 0, 0, errors.New("invalid alias")
	}

	keyType, err := symbiotic.KeyTypeFromString(keyTagParts[1])
	if err != nil {
		return "", 0, 0, err
	}

	keyId, err := strconv.Atoi(keyTagParts[2])
	if err != nil {
		return "", 0, 0, err
	}

	return keyTagParts[0], keyType, keyId, nil
}
