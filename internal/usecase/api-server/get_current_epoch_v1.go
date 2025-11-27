package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetCurrentEpoch handles the gRPC GetCurrentEpoch request
func (h *grpcHandler) GetCurrentEpoch(ctx context.Context, req *apiv1.GetCurrentEpochRequest) (*apiv1.GetCurrentEpochResponse, error) {
	currentEpochInfo, err := h.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Error(codes.NotFound, "no validator set header found")
		}
		return nil, errors.Errorf("failed to get latest validator set header: %w", err)
	}

	return &apiv1.GetCurrentEpochResponse{
		Epoch:     uint64(currentEpochInfo.Epoch),
		StartTime: timestamppb.New(time.Unix(int64(currentEpochInfo.CaptureTimestamp), 0).UTC()),
	}, nil
}
