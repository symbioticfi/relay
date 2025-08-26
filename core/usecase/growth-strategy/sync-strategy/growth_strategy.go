package syncStrategy

import (
	"context"
	"log/slog"
	"math"

	"github.com/ethereum/go-ethereum/common"
	strategyHelpers "github.com/symbioticfi/relay/core/usecase/growth-strategy/helpers"
	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"

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

func (gs GrowthStrategySync) GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, error) {
	lastCommittedAddr, lastCommittedEpoch, err := gs.getLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last committed header epoch: %w", err)
	}

	prevHash, err := gs.client.GetHeaderHash(ctx, lastCommittedAddr)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get last committed header hash for address %s: %w", lastCommittedAddr.Address.Hex(), err)
	}

	if len(config.Replicas) <= 1 {
		return prevHash, nil
	}

	for _, replica := range config.Replicas {
		if replica == lastCommittedAddr {
			continue
		}

		hash, err := gs.client.GetHeaderHashAt(ctx, replica, lastCommittedEpoch)
		if err != nil {
			return common.Hash{}, errors.Errorf("failed to get header hash for replica %d: %w", replica, err)
		}

		if hash != prevHash {
			return common.Hash{}, errors.Errorf("committed headers doesn't match at epoch: %d", lastCommittedEpoch)
		}

		prevHash = hash
	}

	return prevHash, nil
}

func (gs GrowthStrategySync) GetLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (uint64, error) {
	_, epoch, err := gs.getLastCommittedHeaderEpoch(ctx, config)
	return epoch, err
}

func (gs GrowthStrategySync) GetPreviousHash(ctx context.Context, epoch uint64, config entity.NetworkConfig, valset entity.ValidatorSet) (common.Hash, error) {
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

func (gs GrowthStrategySync) GetValsetStatus(ctx context.Context, config entity.NetworkConfig, valset entity.ValidatorSet) (entity.ValidatorSetStatus, error) {
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

func (gs GrowthStrategySync) IsValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	if len(config.Replicas) == 0 {
		return entity.CrossChainAddress{}, false, nil
	}

	for _, addr := range config.Replicas {
		isCommitted, err := gs.client.IsValsetHeaderCommittedAt(ctx, addr, epoch)
		if err != nil {
			return entity.CrossChainAddress{}, false, errors.Errorf("failed to check if valset header is committed at epoch %d: %w", epoch, err)
		}
		if !isCommitted {
			return entity.CrossChainAddress{}, false, nil
		}
	}

	return config.Replicas[0], true, nil
}

func (gs GrowthStrategySync) getLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (entity.CrossChainAddress, uint64, error) {
	minEpoch := uint64(math.MaxUint64)
	var minEpochAddr entity.CrossChainAddress

	for _, addr := range config.Replicas {
		epoch, err := gs.client.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return entity.CrossChainAddress{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch <= minEpoch {
			minEpoch = epoch
			minEpochAddr = addr
		}
	}

	return minEpochAddr, minEpoch, nil
}
