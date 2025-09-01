package growthStrategyTypes

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/core/entity"
)

type GrowthStrategy interface {
	GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, uint64, error)
}
