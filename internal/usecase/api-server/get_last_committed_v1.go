package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/entity"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetLastCommitted handles the gRPC GetLastCommitted request
func (h *grpcHandler) GetLastCommitted(ctx context.Context, req *apiv1.GetLastCommittedRequest) (*apiv1.GetLastCommittedResponse, error) {
	if req.GetSettlementChainId() == 0 {
		return nil, errors.New("settlement chain ID cannot be 0")
	}

	cfg, err := h.cfg.EvmClient.GetConfig(ctx, uint64(time.Now().Unix()))
	if err != nil {
		return nil, errors.Errorf("failed to get config: %w", err)
	}

	var settlementChain *entity.CrossChainAddress

	for _, settlement := range cfg.Settlements {
		if settlement.ChainId == req.GetSettlementChainId() {
			settlementChain = &settlement
			break
		}
	}

	if settlementChain == nil {
		return nil, errors.New("invalid settlement chain ID, not such chain found in network config")
	}

	lastCommittedEpoch, err := h.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, *settlementChain)
	if err != nil {
		return nil, errors.Errorf("failed to get last committed epoch: %w", err)
	}

	// TODO: Get the epoch start time
	epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, lastCommittedEpoch)
	if err != nil {
		return nil, errors.Errorf("failed to get epoch start: %w", err)
	}

	return &apiv1.GetLastCommittedResponse{
		SettlementChainId: req.GetSettlementChainId(),
		EpochInfo: &apiv1.ChainEpochInfo{
			LastCommittedEpoch: lastCommittedEpoch,
			StartTime:          timestamppb.New(time.Unix(int64(epochStart), 0).UTC()),
		},
	}, nil
}
