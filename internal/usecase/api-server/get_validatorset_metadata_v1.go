package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorSetMetadata handles the gRPC GetValidatorSetMetadata request
func (h *grpcHandler) GetValidatorSetMetadata(ctx context.Context, req *apiv1.GetValidatorSetMetadataRequest) (*apiv1.GetValidatorSetMetadataResponse, error) {
	var epochRequested uint64
	if req.Epoch == nil {
		latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
		if err != nil {
			return nil, err
		}

		epochRequested = latestEpoch
	} else {
		epochRequested = req.GetEpoch()
	}

	extraData, commitData, err := h.cfg.Repo.GetValidatorSetMetadata(ctx, epochRequested)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return nil, errors.Errorf("failed to get validator set from epoch: %w", err)
	}

	if errors.Is(err, entity.ErrEntityNotFound) {
		return nil, errors.New("no metadata found for the requested epoch")
	}

	extraDataProto := make([]*apiv1.ExtraData, 0, len(extraData))
	for _, ed := range extraData {
		extraDataProto = append(extraDataProto, &apiv1.ExtraData{
			Key:   ed.Key.Bytes(),
			Value: ed.Value.Bytes(),
		})
	}

	sigRequest := &entity.SignatureRequest{
		KeyTag:        entity.ValsetHeaderKeyTag,
		RequiredEpoch: entity.Epoch(epochRequested),
		Message:       commitData,
	}

	return &apiv1.GetValidatorSetMetadataResponse{
		ExtraData:      extraDataProto,
		CommitmentData: commitData,
		RequestHash:    sigRequest.Hash().Hex(),
	}, nil
}
