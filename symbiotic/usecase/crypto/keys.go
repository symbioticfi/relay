package crypto

import (
	"sync"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/internal/client/repository/cache"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/ecdsaSecp256k1"
)

type PublicKey = symbiotic.PublicKey
type PrivateKey = symbiotic.PrivateKey

var (
	pubKeyCache   cache.Cache[pubKeyCacheKey, symbiotic.PublicKey]
	initCacheOnce sync.Once
)

// InitializePubkeyCache initializes the public key cache with the given size.
// Dev: Should be called only once during application startup.
func InitializePubkeyCache(cacheSize int) (err error) {
	initCacheOnce.Do(func() {
		pubKeyCache, err = cache.NewCache[pubKeyCacheKey, symbiotic.PublicKey](
			cache.Config{Size: cacheSize},
			func(key pubKeyCacheKey) uint32 {
				// FNV-1a 32-bit hash
				const (
					offset32 = uint32(2166136261)
					prime32  = uint32(16777619)
				)
				hash := offset32
				hash ^= uint32(key.keyTag)
				hash *= prime32

				// Process each byte of the key string
				for i := 0; i < len(key.key); i++ {
					hash ^= uint32(key.key[i])
					hash *= prime32
				}
				return hash
			},
		)
	})
	return err
}

type pubKeyCacheKey struct {
	keyTag symbiotic.KeyType
	key    string
}

func NewPublicKey(keyType symbiotic.KeyType, keyBytes symbiotic.RawPublicKey) (PublicKey, error) {
	cacheKey := pubKeyCacheKey{
		keyTag: keyType,
		key:    string(keyBytes),
	}

	if pubKeyCache != nil {
		// Try cache first
		if pk, ok := pubKeyCache.Get(cacheKey); ok {
			return pk, nil
		}
	}

	var (
		pubkey PublicKey
		err    error
	)
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		pubkey, err = blsBn254.FromRaw(keyBytes)
	case symbiotic.KeyTypeEcdsaSecp256k1:
		pubkey, err = ecdsaSecp256k1.FromRaw(keyBytes)
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	default:
		return nil, errors.New("unsupported key type")
	}

	if err != nil {
		return nil, err
	}

	if pubKeyCache != nil {
		pubKeyCache.Add(cacheKey, pubkey)
	}

	return pubkey, nil
}

func NewPrivateKey(keyType symbiotic.KeyType, keyBytes []byte) (PrivateKey, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.NewPrivateKey(keyBytes)
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.NewPrivateKey(keyBytes)
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func HashMessage(keyType symbiotic.KeyType, msg []byte) (symbiotic.RawMessageHash, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.HashMessage(msg), nil
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.HashMessage(msg), nil
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}

func GeneratePrivateKey(keyType symbiotic.KeyType) (PrivateKey, error) {
	switch keyType {
	case symbiotic.KeyTypeBlsBn254:
		return blsBn254.GenerateKey()
	case symbiotic.KeyTypeEcdsaSecp256k1:
		return ecdsaSecp256k1.GenerateKey()
	case symbiotic.KeyTypeInvalid:
		return nil, errors.New("unsupported key type")
	}
	return nil, errors.New("unsupported key type")
}
