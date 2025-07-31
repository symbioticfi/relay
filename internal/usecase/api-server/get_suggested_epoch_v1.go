package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetSuggestedEpoch handles the gRPC GetSuggestedEpoch request
func (h *grpcHandler) GetSuggestedEpoch(ctx context.Context, req *v1.GetSuggestedEpochRequest) (*v1.EpochInfo, error) {
	valset, err := h.cfg.Repo.GetLatestValidatorSet(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set: %w", err)
	}

	// just return latest derived epoch
	// there is no way to make it more deterministic across each node
	return &v1.EpochInfo{
		Epoch:     valset.Epoch,
		StartTime: timestamppb.New(time.Unix(int64(valset.CaptureTimestamp), 0).UTC()),
	}, nil
}
