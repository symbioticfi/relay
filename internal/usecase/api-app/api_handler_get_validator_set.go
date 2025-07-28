package apiApp

import (
	"context"
	"time"

	api2 "github.com/symbioticfi/relay/core/api/gen/api"

	"github.com/go-errors/errors"

	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
)

func (h *handler) GetValidatorSetGet(ctx context.Context, params api2.GetValidatorSetGetParams) (*api2.ValidatorSet, error) {
	latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}

	epochRequested := params.Epoch.Value
	if !params.Epoch.IsSet() {
		epochRequested = latestEpoch
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

	return convertValidatorSetToAPI(validatorSet), nil
}

func convertValidatorSetToAPI(valSet entity.ValidatorSet) *api2.ValidatorSet {
	return &api2.ValidatorSet{
		Version:            valSet.Version,
		RequiredKeyTag:     uint8(valSet.RequiredKeyTag),
		Epoch:              valSet.Epoch,
		CaptureTimestamp:   time.Unix(int64(valSet.CaptureTimestamp), 0).UTC(),
		QuorumThreshold:    valSet.QuorumThreshold.String(),
		PreviousHeaderHash: valSet.PreviousHeaderHash.Hex(),
		Status:             uint8(valSet.Status),
		Validators: lo.Map(valSet.Validators, func(v entity.Validator, _ int) api2.Validator {
			return api2.Validator{
				Operator:    v.Operator.Hex(),
				VotingPower: v.VotingPower.String(),
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k entity.ValidatorKey, _ int) api2.Key {
					return api2.Key{
						Tag:     uint8(k.Tag),
						Payload: k.Payload,
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v entity.ValidatorVault, _ int) api2.ValidatorVault {
					return api2.ValidatorVault{
						ChainId:     v.ChainID,
						Vault:       v.Vault.Hex(),
						VotingPower: v.VotingPower.String(),
					}
				}),
			}
		}),
	}
}
