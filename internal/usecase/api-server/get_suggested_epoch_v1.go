package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetSuggestedEpoch handles the gRPC GetSuggestedEpoch request
func (h *grpcHandler) GetSuggestedEpoch(ctx context.Context, req *apiv1.GetSuggestedEpochRequest) (*apiv1.GetSuggestedEpochResponse, error) {
	valset, err := h.cfg.Repo.GetLatestValidatorSetMeta(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set: %w", err)
	}

	// just return latest derived epoch
	// there is no way to make it more deterministic across each node
	return &apiv1.GetSuggestedEpochResponse{
		Epoch:     valset.Epoch,
		StartTime: timestamppb.New(time.Unix(int64(valset.CaptureTimestamp), 0).UTC()),
	}, nil
}
