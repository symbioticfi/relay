package growthStrategy

import (
	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	newestStrategy "github.com/symbioticfi/relay/core/usecase/growth-strategy/async-strategy"
	noSettlementStrategy "github.com/symbioticfi/relay/core/usecase/growth-strategy/no-settlement-strategy"
	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"
	syncStrategy "github.com/symbioticfi/relay/core/usecase/growth-strategy/sync-strategy"
)

func NewGrowthStrategy(gst entity.GrowthStrategyType, client evm.IEvmClient) (strategyTypes.GrowthStrategy, error) {
	switch gst {
	case entity.GrowthStrategyAsync:
		return newestStrategy.NewGrowthStrategyAsync(client), nil
	case entity.GrowthStrategySync:
		return syncStrategy.NewGrowthStrategySync(client), nil
	case entity.GrowthStrategyNoSettlement:
		return noSettlementStrategy.NewGrowthStrategyNoSettlement(client), nil
	}

	return nil, errors.New("unknown growth strategy")
}
