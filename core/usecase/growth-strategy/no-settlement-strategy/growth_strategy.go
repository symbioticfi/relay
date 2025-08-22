package noSettlementStrategy

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/symbioticfi/relay/core/client/evm"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/entity"
)

var NoSettlementHash = common.HexToHash("NoSettlementHash")

type GrowthStrategyNoSettlement struct {
	client evm.IEvmClient
}

func NewGrowthStrategyNoSettlement(client evm.IEvmClient) *GrowthStrategyNoSettlement {
	return &GrowthStrategyNoSettlement{client: client}
}

func (gs GrowthStrategyNoSettlement) GetLastCommittedHeaderHash(ctx context.Context, _ entity.NetworkConfig) (common.Hash, uint64, error) {
	epoch, err := gs.client.GetCurrentEpoch(ctx)
	if err != nil {
		return common.Hash{}, 0, errors.Errorf("failed to get current epoch: %s", err)
	}

	return NoSettlementHash, epoch, nil
}
