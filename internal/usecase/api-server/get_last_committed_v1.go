package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetLastCommitted handles the gRPC GetLastCommitted request
func (h *grpcHandler) GetLastCommitted(ctx context.Context, req *apiv1.GetLastCommittedRequest) (*apiv1.GetLastCommittedResponse, error) {
	if req.GetSettlementChainId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "settlement chain ID cannot be 0")
	}

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

	var settlementChain *symbiotic.CrossChainAddress

	for _, settlement := range cfg.Settlements {
		if settlement.ChainId == req.GetSettlementChainId() {
			settlementChain = &settlement
			break
		}
	}

	if settlementChain == nil {
		return nil, status.Errorf(codes.NotFound, "settlement chain ID %d not found in network config", req.GetSettlementChainId())
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
			LastCommittedEpoch: uint64(lastCommittedEpoch),
			StartTime:          timestamppb.New(time.Unix(int64(epochStart), 0).UTC()),
		},
	}, nil
}
