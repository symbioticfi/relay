package growthStrategy

import (
	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	newestStrategy "github.com/symbioticfi/relay/core/usecase/growth-strategy/newest-strategy"
	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"
	syncStrategy "github.com/symbioticfi/relay/core/usecase/growth-strategy/sync-strategy"
)

func NewGrowthStrategy(gst entity.GrowthStrategyType, client *evm.Client) (strategyTypes.GrowthStrategy, error) {
	switch gst {
	case entity.GrowthStrategyNewest:
		return newestStrategy.NewGrowthStrategyNewest(client), nil
	case entity.GrowthStrategySync:
		return syncStrategy.NewGrowthStrategySync(client), nil
	}

	return nil, errors.New("unknown growth strategy")
}
