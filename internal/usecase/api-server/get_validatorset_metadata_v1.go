package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorSetMetadata handles the gRPC GetValidatorSetMetadata request
func (h *grpcHandler) GetValidatorSetMetadata(ctx context.Context, req *apiv1.GetValidatorSetMetadataRequest) (*apiv1.GetValidatorSetMetadataResponse, error) {
	var epochRequested entity.Epoch
	if req.Epoch == nil {
		latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
		if err != nil {
			return nil, err
		}

		epochRequested = latestEpoch
	} else {
		epochRequested = entity.Epoch(req.GetEpoch())
	}

	metadata, err := h.cfg.Repo.GetValidatorSetMetadata(ctx, epochRequested)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, errors.Errorf("no metadata found for the requested epoch: %w", err)
		}
		return nil, errors.Errorf("failed to get validator set from epoch: %w", err)
	}

	return &apiv1.GetValidatorSetMetadataResponse{
		RequestId: metadata.RequestID.Hex(),
		ExtraData: lo.Map(metadata.ExtraData, func(ed entity.ExtraData, _ int) *apiv1.ExtraData {
			return &apiv1.ExtraData{
				Key:   ed.Key.Bytes(),
				Value: ed.Value.Bytes(),
			}
		}),
		CommitmentData: metadata.CommitmentData,
	}, nil
}
