package api_server

import (
	"bytes"
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetValidatorByKey handles the gRPC GetValidatorByKey request
func (h *grpcHandler) GetValidatorByKey(ctx context.Context, req *apiv1.GetValidatorByKeyRequest) (*apiv1.GetValidatorByKeyResponse, error) {
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

	keyTag := req.GetKeyTag()
	if keyTag <= 0 {
		return nil, status.Error(codes.InvalidArgument, "key tag must be positive")
	}

	onChainKey := req.GetOnChainKey()
	if len(onChainKey) == 0 {
		return nil, status.Error(codes.InvalidArgument, "on chain key is empty")
	}

	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	validator, found := lo.Find(validatorSet.Validators, func(validator symbiotic.Validator) bool {
		_, found := lo.Find(validator.Keys, func(key symbiotic.ValidatorKey) bool {
			return key.Tag == symbiotic.KeyTag(keyTag) && bytes.Equal(key.Payload, onChainKey)
		})
		return found
	})
	if !found {
		return nil, status.Error(codes.NotFound, "validator not found")
	}

	return &apiv1.GetValidatorByKeyResponse{Validator: convertValidatorToPB(validator)}, nil
}
