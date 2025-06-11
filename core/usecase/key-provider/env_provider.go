package keyprovider

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"sync"

	"middleware-offchain/core/entity"
)

type EnvKeyProvider struct {
	cache map[entity.KeyTag][]byte
	mu    sync.RWMutex
}

func NewEnvKeyProvider() *EnvKeyProvider {
	return &EnvKeyProvider{
		cache: make(map[entity.KeyTag][]byte),
	}
}

func (e *EnvKeyProvider) GetPrivateKey(keyTag entity.KeyTag) ([]byte, error) {
	e.mu.RLock()

	key, ok := e.cache[keyTag]
	if ok {
		e.mu.RUnlock()
		return key, nil
	}
	e.mu.RUnlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	val := os.Getenv(strings.ToUpper(alias)) // todo ilya: research if it's safe to read private keys from environment variables
	if val == "" {
		return nil, errors.New("key not found in environment")
	}

	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.cache[keyTag] = decoded
	e.mu.Unlock()

	return decoded, nil
}

func (e *EnvKeyProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	e.mu.RLock()
	_, ok := e.cache[keyTag]
	e.mu.RUnlock()

	if ok {
		return true, nil
	}

	pk, err := e.GetPrivateKey(keyTag)
	if err != nil {
		return false, err
	}

	return len(pk) > 0, nil
}
