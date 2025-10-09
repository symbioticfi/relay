package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetValidatorSet handles the gRPC GetValidatorSet request
func (h *grpcHandler) GetValidatorSet(ctx context.Context, req *apiv1.GetValidatorSetRequest) (*apiv1.GetValidatorSetResponse, error) {
	latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set epoch: %w", err)
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = symbiotic.Epoch(req.GetEpoch())
	}

	// epoch from future
	if epochRequested > latestEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "epoch %d is greater than latest epoch %d", epochRequested, latestEpoch)
	}

	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	return convertValidatorSetToPB(validatorSet), nil
}

// getValidatorSetForEpoch retrieves validator set for a given epoch, either from repo or by deriving it
func (h *grpcHandler) getValidatorSetForEpoch(ctx context.Context, epochRequested symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	validatorSet, err := h.cfg.Repo.GetValidatorSetByEpoch(ctx, epochRequested)
	if err == nil {
		return validatorSet, nil
	}

	if errors.Is(err, entity.ErrEntityNotFound) {
		return symbiotic.ValidatorSet{}, status.Errorf(codes.NotFound, "validator set for epoch %d not found", epochRequested)
	}

	return symbiotic.ValidatorSet{}, errors.Errorf("failed to get validator set for epoch %d: %w", epochRequested, err)
}

func convertValidatorSetToPB(valSet symbiotic.ValidatorSet) *apiv1.GetValidatorSetResponse {
	return &apiv1.GetValidatorSetResponse{
		Version:          uint32(valSet.Version),
		RequiredKeyTag:   uint32(valSet.RequiredKeyTag),
		Epoch:            uint64(valSet.Epoch),
		CaptureTimestamp: timestamppb.New(time.Unix(int64(valSet.CaptureTimestamp), 0).UTC()),
		QuorumThreshold:  valSet.QuorumThreshold.String(),
		Status:           convertValidatorSetStatusToPB(valSet.Status),
		Validators: lo.Map(valSet.Validators, func(v symbiotic.Validator, _ int) *apiv1.Validator {
			return convertValidatorToPB(v)
		}),
	}
}

func convertValidatorToPB(v symbiotic.Validator) *apiv1.Validator {
	return &apiv1.Validator{
		Operator:    v.Operator.Hex(),
		VotingPower: v.VotingPower.String(),
		IsActive:    v.IsActive,
		Keys: lo.Map(v.Keys, func(k symbiotic.ValidatorKey, _ int) *apiv1.Key {
			return &apiv1.Key{
				Tag:     uint32(k.Tag),
				Payload: k.Payload,
			}
		}),
		Vaults: lo.Map(v.Vaults, func(v symbiotic.ValidatorVault, _ int) *apiv1.ValidatorVault {
			return &apiv1.ValidatorVault{
				ChainId:     v.ChainID,
				Vault:       v.Vault.Hex(),
				VotingPower: v.VotingPower.String(),
			}
		}),
	}
}

func convertValidatorSetStatusToPB(status symbiotic.ValidatorSetStatus) apiv1.ValidatorSetStatus {
	switch status {
	case symbiotic.HeaderDerived:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_DERIVED
	case symbiotic.HeaderAggregated:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_AGGREGATED
	case symbiotic.HeaderCommitted:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_COMMITTED
	default:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_UNSPECIFIED
	}
}
