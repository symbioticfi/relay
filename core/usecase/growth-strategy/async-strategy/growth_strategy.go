package newestStrategy

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
)

type GrowthStrategyAsync struct {
	client evm.IEvmClient
}

func NewGrowthStrategyAsync(client evm.IEvmClient) *GrowthStrategyAsync {
	return &GrowthStrategyAsync{client: client}
}

func (gs GrowthStrategyAsync) GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, uint64, error) {
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

	hash, err := gs.client.GetHeaderHash(ctx, maxEpochAddr)
	if err != nil {
		return common.Hash{}, 0, errors.Errorf("failed to get last committed header hash for address %s: %w", maxEpochAddr.Address.Hex(), err)
	}

	return hash, maxEpoch, nil
}
