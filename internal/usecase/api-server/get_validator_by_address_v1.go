package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorByAddress handles the gRPC GetValidatorByAddress request
func (h *grpcHandler) GetValidatorByAddress(ctx context.Context, req *apiv1.GetValidatorByAddressRequest) (*apiv1.GetValidatorByAddressResponse, error) {
	latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = req.GetEpoch()
	}

	// epoch from future
	if epochRequested > latestEpoch {
		return nil, errors.New("epoch requested is greater than latest epoch")
	}

	// parse validator address
	if !common.IsHexAddress(req.GetAddress()) {
		return nil, errors.New("invalid validator address format")
	}
	validatorAddress := common.HexToAddress(req.GetAddress())

	// get validator set for the epoch
	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	// find validator by address
	validator, found := lo.Find(validatorSet.Validators, func(v entity.Validator) bool {
		return v.Operator == validatorAddress
	})
	if !found {
		return nil, errors.New("validator not found for the given address and epoch")
	}

	return &apiv1.GetValidatorByAddressResponse{
		Validator: convertValidatorToPB(validator),
	}, nil
}
