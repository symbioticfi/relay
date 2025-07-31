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
	client *evm.Client
}

func NewGrowthStrategySync(client *evm.Client) *GrowthStrategySync {
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

	if config.MaxMissingEpochs != 0 {
		currentEpoch, err := gs.client.GetCurrentEpoch(ctx)
		if err != nil {
			return common.Hash{}, 0, errors.Errorf("failed to get current epoch: %w", err)
		}

		if currentEpoch-minEpoch < config.MaxMissingEpochs {
			return common.Hash{}, 0, errors.New("max missing epochs exceeded")
		}
	}

	hash, err := gs.client.GetHeaderHash(ctx, minEpochAddr)
	if err != nil {
		return common.Hash{}, 0, errors.Errorf("failed to get last committed header hash for address %s: %w", minEpochAddr.Address.Hex(), err)
	}

	return hash, minEpoch, nil
}
