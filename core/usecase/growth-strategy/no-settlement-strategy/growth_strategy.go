package noSettlementStrategy

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/symbioticfi/relay/core/client/evm"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/entity"
)

var NoSettlementAddr = entity.CrossChainAddress{ChainId: 111111111111}

type GrowthStrategyNoSettlement struct {
	client evm.IEvmClient
}

func NewGrowthStrategyNoSettlement(client evm.IEvmClient) *GrowthStrategyNoSettlement {
	return &GrowthStrategyNoSettlement{client: client}
}

func (gs GrowthStrategyNoSettlement) GetLastCommittedHeaderHash(_ context.Context, _ entity.NetworkConfig) (common.Hash, error) {
	return common.Hash{}, nil
}

func (gs GrowthStrategyNoSettlement) GetLastCommittedHeaderEpoch(ctx context.Context, _ entity.NetworkConfig) (uint64, error) {
	epoch, err := gs.client.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, errors.Errorf("failed to get current epoch: %s", err)
	}
	return epoch, nil
}

func (gs GrowthStrategyNoSettlement) GetValsetStatus(ctx context.Context, _ entity.NetworkConfig, valset entity.ValidatorSet) (entity.ValidatorSetStatus, error) {
	currentEpoch, err := gs.client.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, errors.Errorf("failed to get current epoch: %w", err)
	}

	if valset.Epoch != currentEpoch { // TODO: policy when no settlement
		return entity.HeaderInactive, nil
	}

	return entity.HeaderActive, nil
}

func (gs GrowthStrategyNoSettlement) IsValsetHeaderCommitted(_ context.Context, _ entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	return NoSettlementAddr, true, nil
}
