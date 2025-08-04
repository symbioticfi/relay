package syncStrategy

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
)

type GrowthStrategySync struct {
	client evm.IEvmClient
}

func NewGrowthStrategySync(client evm.IEvmClient) *GrowthStrategySync {
	return &GrowthStrategySync{client: client}
}

func (gs GrowthStrategySync) GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, uint64, error) {
	minEpoch := uint64(math.MaxUint64)
	var minEpochAddr entity.CrossChainAddress

	for _, addr := range config.Replicas {
		epoch, err := gs.client.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return common.Hash{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch <= minEpoch {
			minEpoch = epoch
			minEpochAddr = addr
		}
	}

	prevHash, err := gs.client.GetHeaderHash(ctx, minEpochAddr)
	if err != nil {
		return common.Hash{}, 0, errors.Errorf("failed to get last committed header hash for address %s: %w", minEpochAddr.Address.Hex(), err)
	}

	if len(config.Replicas) <= 1 {
		return prevHash, minEpoch, nil
	}

	for _, replica := range config.Replicas {
		if replica == minEpochAddr {
			continue
		}

		hash, err := gs.client.GetHeaderHashAt(ctx, replica, minEpoch)
		if err != nil {
			return common.Hash{}, 0, errors.Errorf("failed to get header hash for replica %d: %w", replica, err)
		}

		if hash != prevHash {
			return common.Hash{}, 0, errors.Errorf("committed headers doesn't match at epoch: %d", minEpoch)
		}

		prevHash = hash
	}

	return prevHash, minEpoch, nil
}
