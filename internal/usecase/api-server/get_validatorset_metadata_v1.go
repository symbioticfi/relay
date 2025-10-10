package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetValidatorSetMetadata handles the gRPC GetValidatorSetMetadata request
func (h *grpcHandler) GetValidatorSetMetadata(ctx context.Context, req *apiv1.GetValidatorSetMetadataRequest) (*apiv1.GetValidatorSetMetadataResponse, error) {
	var epochRequested symbiotic.Epoch
	if req.Epoch == nil {
		latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to get latest validator set epoch: %w", err)
		}
		epochRequested = latestEpoch
	} else {
		epochRequested = symbiotic.Epoch(req.GetEpoch())
	}

	metadata, err := h.cfg.Repo.GetValidatorSetMetadata(ctx, epochRequested)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "no metadata found for epoch %d", epochRequested)
		}
		return nil, errors.Errorf("failed to get validator set metadata from epoch: %w", err)
	}

	return &apiv1.GetValidatorSetMetadataResponse{
		RequestId: metadata.RequestID.Hex(),
		ExtraData: lo.Map(metadata.ExtraData, func(ed symbiotic.ExtraData, _ int) *apiv1.ExtraData {
			return &apiv1.ExtraData{
				Key:   ed.Key.Bytes(),
				Value: ed.Value.Bytes(),
			}
		}),
		CommitmentData: metadata.CommitmentData,
	}, nil
}
