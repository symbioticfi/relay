package newestStrategy

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
)

type GrowthStrategyNewest struct {
	client *evm.Client
}

func NewGrowthStrategyNewest(client *evm.Client) *GrowthStrategyNewest {
	return &GrowthStrategyNewest{client: client}
}

func (gs GrowthStrategyNewest) GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, uint64, error) {
	maxEpoch := uint64(0)
	var maxEpochAddr entity.CrossChainAddress

	for _, addr := range config.Replicas {
		epoch, err := gs.client.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return common.Hash{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch >= maxEpoch {
			maxEpoch = epoch
			maxEpochAddr = addr
		}
	}

	if config.MaxMissingEpochs != 0 {
		currentEpoch, err := gs.client.GetCurrentEpoch(ctx)
		if err != nil {
			return common.Hash{}, 0, errors.Errorf("failed to get current epoch: %w", err)
		}

		if currentEpoch-maxEpoch < config.MaxMissingEpochs {
			return common.Hash{}, 0, errors.New("max missing epochs exceeded")
		}
	}

	hash, err := gs.client.GetHeaderHash(ctx, maxEpochAddr)
	if err != nil {
		return common.Hash{}, 0, errors.Errorf("failed to get last committed header hash for address %s: %w", maxEpochAddr.Address.Hex(), err)
	}

	return hash, maxEpoch, nil
}
