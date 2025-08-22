package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorSet handles the gRPC GetValidatorSet request
func (h *grpcHandler) GetValidatorSet(ctx context.Context, req *apiv1.GetValidatorSetRequest) (*apiv1.GetValidatorSetResponse, error) {
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

	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	return convertValidatorSetToPB(validatorSet), nil
}

// getValidatorSetForEpoch retrieves validator set for a given epoch, either from repo or by deriving it
func (h *grpcHandler) getValidatorSetForEpoch(ctx context.Context, epochRequested uint64) (entity.ValidatorSet, error) {
	validatorSet, err := h.cfg.Repo.GetValidatorSetByEpoch(ctx, epochRequested)
	if err == nil {
		return validatorSet, nil
	}
	if !errors.Is(err, entity.ErrEntityNotFound) {
		return entity.ValidatorSet{}, errors.Errorf("failed to get validator set for epoch %d: %v", epochRequested, err)
	}

	// if error it means that epoch is not derived / committed yet
	// so we need to derive it
	epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, epochRequested)
	if err != nil {
		return entity.ValidatorSet{}, err
	}
	config, err := h.cfg.EvmClient.GetConfig(ctx, epochStart)
	if err != nil {
		return entity.ValidatorSet{}, err
	}
	validatorSet, err = h.cfg.Deriver.GetValidatorSet(ctx, epochRequested, config)
	if err != nil {
		return entity.ValidatorSet{}, err
	}
	return validatorSet, nil
}

func convertValidatorSetToPB(valSet entity.ValidatorSet) *apiv1.GetValidatorSetResponse {
	return &apiv1.GetValidatorSetResponse{
		Version:            uint32(valSet.Version),
		RequiredKeyTag:     uint32(valSet.RequiredKeyTag),
		Epoch:              valSet.Epoch,
		CaptureTimestamp:   timestamppb.New(time.Unix(int64(valSet.CaptureTimestamp), 0).UTC()),
		QuorumThreshold:    valSet.QuorumThreshold.String(),
		PreviousHeaderHash: valSet.PreviousHeaderHash.Hex(),
		Status:             convertValidatorSetStatusToPB(valSet.Status),
		Validators: lo.Map(valSet.Validators, func(v entity.Validator, _ int) *apiv1.Validator {
			return convertValidatorToPB(v)
		}),
	}
}

func convertValidatorToPB(v entity.Validator) *apiv1.Validator {
	return &apiv1.Validator{
		Operator:    v.Operator.Hex(),
		VotingPower: v.VotingPower.String(),
		IsActive:    v.IsActive,
		Keys: lo.Map(v.Keys, func(k entity.ValidatorKey, _ int) *apiv1.Key {
			return &apiv1.Key{
				Tag:     uint32(k.Tag),
				Payload: k.Payload,
			}
		}),
		Vaults: lo.Map(v.Vaults, func(v entity.ValidatorVault, _ int) *apiv1.ValidatorVault {
			return &apiv1.ValidatorVault{
				ChainId:     v.ChainID,
				Vault:       v.Vault.Hex(),
				VotingPower: v.VotingPower.String(),
			}
		}),
	}
}

func convertValidatorSetStatusToPB(status entity.ValidatorSetStatus) apiv1.ValidatorSetStatus {
	switch status {
	case entity.HeaderInactive:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_INACTIVE
	case entity.HeaderActive:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_ACTIVE
	default:
		return apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_UNSPECIFIED
	}
}
