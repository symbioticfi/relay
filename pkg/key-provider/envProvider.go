package keyprovider

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

type EnvKeyProvider struct {
	cache map[uint8][]byte
}

func NewEnvKeyProvider() *EnvKeyProvider {
	return &EnvKeyProvider{}
}

func (e *EnvKeyProvider) GetPrivateKey(keyTag uint8) ([]byte, error) {
	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}
	val := os.Getenv(strings.ToUpper(alias))
	if val == "" {
		return nil, errors.New("key not found in environment")
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}
	e.cache[keyTag] = decoded
	return decoded, nil
}

func (e *EnvKeyProvider) HasKey(keyTag uint8) (bool, error) {
	_, ok := e.cache[keyTag]
	if ok {
		return true, nil
	}

	pk, err := e.GetPrivateKey(keyTag)
	if err != nil {
		return false, err
	}
	return len(pk) > 0, nil
}
