package strategyTypes

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/core/entity"
)

var EmptyValsetHeaderHash = common.HexToHash("0x868e09d528a16744c1f38ea3c10cc2251e01a456434f91172247695087d129b7")

type GrowthStrategy interface {
	GetLastCommittedHeaderHash(ctx context.Context, config entity.NetworkConfig) (common.Hash, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (uint64, error)
	GetPreviousHash(ctx context.Context, epoch uint64, config entity.NetworkConfig, valset entity.ValidatorSet) (common.Hash, error)
	GetValsetStatus(ctx context.Context, config entity.NetworkConfig, valset entity.ValidatorSet) (entity.ValidatorSetStatus, error)
	IsValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error)
}
