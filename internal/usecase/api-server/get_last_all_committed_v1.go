package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/internal/entity"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetLastAllCommitted handles the gRPC GetLastAllCommitted request
func (h *grpcHandler) GetLastAllCommitted(ctx context.Context, _ *apiv1.GetLastAllCommittedRequest) (*apiv1.GetLastAllCommittedResponse, error) {
	currentEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		if !errors.Is(err, entity.ErrEntityNotFound) {
			return nil, errors.Errorf("failed to get current epoch: %w", err)
		}
	}

	cfg, err := h.cfg.EvmClient.GetConfig(ctx, symbiotic.Timestamp(uint64(time.Now().Unix())), currentEpoch)
	if err != nil {
		return nil, errors.Errorf("failed to get config: %w", err)
	}

	epochInfos := make(map[uint64]*apiv1.ChainEpochInfo)
	for _, chain := range cfg.Settlements {
		lastCommittedEpoch, err := h.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, chain)
		if err != nil {
			return nil, errors.Errorf("failed to get last committed epoch for chain %d: %w", chain.ChainId, err)
		}

		epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, lastCommittedEpoch)
		if err != nil {
			return nil, errors.Errorf("failed to get epoch start for chain %d: %w", chain.ChainId, err)
		}

		epochInfos[chain.ChainId] = &apiv1.ChainEpochInfo{
			LastCommittedEpoch: uint64(lastCommittedEpoch),
			StartTime:          timestamppb.New(time.Unix(int64(epochStart), 0).UTC()),
		}
	}

	return &apiv1.GetLastAllCommittedResponse{
		EpochInfos: epochInfos,
	}, nil
}
