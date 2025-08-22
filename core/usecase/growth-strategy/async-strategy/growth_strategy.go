package newestStrategy

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	strategyHelpers "github.com/symbioticfi/relay/core/usecase/growth-strategy/helpers"
	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"

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

func (gs GrowthStrategyAsync) GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, error) {
	lastCommittedAddr, _, err := gs.getLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last committed header epoch: %w", err)
	}

	hash, err := gs.client.GetHeaderHash(ctx, lastCommittedAddr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last committed header hash for address %s: %w", lastCommittedAddr.Address.Hex(), err)
	}

	return hash, nil
}

func (gs GrowthStrategyAsync) GetLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (uint64, error) {
	_, epoch, err := gs.getLastCommittedHeaderEpoch(ctx, config)
	return epoch, err
}

func (gs GrowthStrategyAsync) GetPreviousHash(ctx context.Context, epoch uint64, config entity.NetworkConfig, valset entity.ValidatorSet) (common.Hash, error) {
	committedAddr, isValsetCommitted, err := gs.IsValsetHeaderCommitted(ctx, config, epoch)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to check if validator committed at epoch %d: %w", epoch, err)
	}

	if isValsetCommitted {
		return strategyHelpers.GetPreviousHashForCommittedValset(ctx, gs.client, committedAddr, epoch, valset)
	}

	latestCommittedEpoch, err := gs.GetLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last commmitted valset epoch: %w", err)
	}
	// valset not committed

	if epoch < latestCommittedEpoch {
		slog.DebugContext(ctx, "Header is not committed [missed header]", "epoch", epoch)
		// zero PreviousHeaderHash cos header is orphaned
		return strategyTypes.EmptyValsetHeaderHash, nil
	}

	lastCommittedHash, err := gs.GetLastCommittedHeaderHash(ctx, config)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last committed valset hash: %w", err)
	}

	// trying to link to latest committed header
	slog.DebugContext(ctx, "Header is not committed [new header]", "epoch", epoch)

	return lastCommittedHash, nil
}

func (gs GrowthStrategyAsync) GetValsetStatus(ctx context.Context, config entity.NetworkConfig, valset entity.ValidatorSet) (entity.ValidatorSetStatus, error) {
	currentEpoch, err := gs.client.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, errors.Errorf("failed to get current epoch: %w", err)
	}

	committedEpoch, err := gs.GetLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return 0, errors.Errorf("failed to get last committed header epoch: %w", err)
	}

	if (config.MaxMissingEpochs != 0 && currentEpoch-valset.Epoch > config.MaxMissingEpochs) || valset.Epoch != committedEpoch {
		return entity.HeaderInactive, nil
	}

	return entity.HeaderActive, nil
}

func (gs GrowthStrategyAsync) IsValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	for _, addr := range config.Replicas {
		isCommitted, err := gs.client.IsValsetHeaderCommittedAt(ctx, addr, epoch)
		if err != nil {
			return entity.CrossChainAddress{}, false, errors.Errorf("failed to check if valset header is committed at epoch %d: %w", epoch, err)
		}
		if isCommitted {
			return addr, true, nil
		}
	}
	return entity.CrossChainAddress{}, false, nil
}

func (gs GrowthStrategyAsync) getLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (entity.CrossChainAddress, uint64, error) {
	maxEpoch := uint64(0)
	var maxEpochAddr entity.CrossChainAddress

	for _, addr := range config.Replicas {
		epoch, err := gs.client.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return entity.CrossChainAddress{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch >= maxEpoch {
			maxEpoch = epoch
			maxEpochAddr = addr
		}
	}

	return maxEpochAddr, maxEpoch, nil
}
