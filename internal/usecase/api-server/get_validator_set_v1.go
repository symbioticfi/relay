package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorSet handles the gRPC GetValidatorSet request
func (h *grpcHandler) GetValidatorSet(ctx context.Context, req *v1.GetValidatorSetRequest) (*v1.ValidatorSet, error) {
	latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = *req.Epoch
	}

	// epoch from future
	if epochRequested > latestEpoch {
		return nil, errors.New("epoch requested is greater than latest epoch")
	}

	validatorSet, err := h.cfg.Repo.GetValidatorSetByEpoch(ctx, epochRequested)

	// if error it means that epoch is not derived / committed yet
	// so we need to derive it
	if err != nil {
		epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, epochRequested)
		if err != nil {
			return nil, err
		}
		config, err := h.cfg.EvmClient.GetConfig(ctx, epochRequested)
		if err != nil {
			return nil, err
		}
		validatorSet, err = h.cfg.Deriver.GetValidatorSet(ctx, epochStart, config)
		if err != nil {
			return nil, err
		}
	}

	return convertValidatorSetToPB(validatorSet), nil
}

func convertValidatorSetToPB(valSet entity.ValidatorSet) *v1.ValidatorSet {
	return &v1.ValidatorSet{
		Version:            uint32(valSet.Version),
		RequiredKeyTag:     uint32(valSet.RequiredKeyTag),
		Epoch:              valSet.Epoch,
		CaptureTimestamp:   timestamppb.New(time.Unix(int64(valSet.CaptureTimestamp), 0).UTC()),
		QuorumThreshold:    valSet.QuorumThreshold.String(),
		PreviousHeaderHash: valSet.PreviousHeaderHash.Hex(),
		Status:             convertValidatorSetStatusToPB(valSet.Status),
		Validators: lo.Map(valSet.Validators, func(v entity.Validator, _ int) *v1.Validator {
			return &v1.Validator{
				Operator:    v.Operator.Hex(),
				VotingPower: v.VotingPower.String(),
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k entity.ValidatorKey, _ int) *v1.Key {
					return &v1.Key{
						Tag:     uint32(k.Tag),
						Payload: k.Payload,
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v entity.ValidatorVault, _ int) *v1.ValidatorVault {
					return &v1.ValidatorVault{
						ChainId:     v.ChainID,
						Vault:       v.Vault.Hex(),
						VotingPower: v.VotingPower.String(),
					}
				}),
			}
		}),
	}
}

func convertValidatorSetStatusToPB(status entity.ValidatorSetStatus) v1.ValidatorSetStatus {
	switch status {
	case entity.HeaderPending:
		return v1.ValidatorSetStatus_VALIDATOR_SET_STATUS_PENDING
	case entity.HeaderMissed:
		return v1.ValidatorSetStatus_VALIDATOR_SET_STATUS_MISSED
	case entity.HeaderCommitted:
		return v1.ValidatorSetStatus_VALIDATOR_SET_STATUS_COMMITTED
	default:
		return v1.ValidatorSetStatus_VALIDATOR_SET_STATUS_UNSPECIFIED
	}
}
