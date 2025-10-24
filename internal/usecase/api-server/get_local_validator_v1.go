package api_server

import (
	"context"

	"github.com/go-errors/errors"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetLocalValidator handles the gRPC GetLocalValidator request
func (h *grpcHandler) GetLocalValidator(ctx context.Context, req *apiv1.GetLocalValidatorRequest) (*apiv1.GetLocalValidatorResponse, error) {
	latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set epoch: %w", err)
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = symbiotic.Epoch(req.GetEpoch())
	}

	if epochRequested > latestEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "epoch %d is greater than latest epoch %d", epochRequested, latestEpoch)
	}

	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	pubkey, err := h.cfg.KeyProvider.GetOnchainKeyFromCache(symbiotic.ValsetHeaderKeyTag)
	if err != nil {
		return nil, errors.Errorf("failed to get onchain key from cache: %w", err)
	}

	validator, found := validatorSet.FindValidatorByKey(symbiotic.ValsetHeaderKeyTag, pubkey)
	if !found {
		return nil, status.Errorf(codes.NotFound, "local validator not found")
	}

	return &apiv1.GetLocalValidatorResponse{Validator: convertValidatorToPB(validator)}, nil
}
