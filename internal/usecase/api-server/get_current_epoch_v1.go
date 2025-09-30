package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetCurrentEpoch handles the gRPC GetCurrentEpoch request
func (h *grpcHandler) GetCurrentEpoch(ctx context.Context, req *apiv1.GetCurrentEpochRequest) (*apiv1.GetCurrentEpochResponse, error) {
	currentEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get current epoch: %w", err)
	}

	epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, currentEpoch)
	if err != nil {
		return nil, errors.Errorf("failed to get epoch start: %w", err)
	}

	return &apiv1.GetCurrentEpochResponse{
		Epoch:     uint64(currentEpoch),
		StartTime: timestamppb.New(time.Unix(int64(epochStart), 0).UTC()),
	}, nil
}
