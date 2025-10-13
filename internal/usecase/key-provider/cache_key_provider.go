package keyprovider

import (
	"sync"

	"github.com/symbioticfi/relay/symbiotic/entity"
)

type CacheKeyProvider struct {
	KeyProvider

	nodeKeyMap *sync.Map // map[keyTag]CompactPublicKey
}

func NewCacheKeyProvider(kp KeyProvider) *CacheKeyProvider {
	return &CacheKeyProvider{
		KeyProvider: kp,
		nodeKeyMap:  &sync.Map{},
	}
}

func (c *CacheKeyProvider) GetOnchainKeyFromCache(keyTag entity.KeyTag) (entity.CompactPublicKey, error) {
	onchainKey, ok := c.nodeKeyMap.Load(keyTag)
	if !ok {
		symbPrivate, err := c.GetPrivateKey(keyTag)
		if err != nil {
			return nil, err
		}

		onchainKey = symbPrivate.PublicKey().OnChain()
		c.nodeKeyMap.Store(keyTag, onchainKey)
	}
	return onchainKey.(entity.CompactPublicKey), nil
}
