package apiApp

import (
	"context"
	"time"

	"github.com/samber/lo"

	"middleware-offchain/core/entity"
	"middleware-offchain/internal/gen/api"
)

func (h *handler) GetValidatorSetGet(ctx context.Context, params api.GetValidatorSetGetParams) (*api.ValidatorSet, error) {
	epoch := params.Epoch.Value
	if !params.Epoch.IsSet() {
		var err error
		epoch, err = h.cfg.EVMClient.GetCurrentEpoch(ctx)
		if err != nil {
			return nil, err
		}
	}

	validatorSet, err := h.cfg.Repo.GetValidatorSetByEpoch(ctx, epoch)
	if err != nil {
		return nil, err
	}

	return convertValidatorSetToAPI(validatorSet), nil
}

func convertValidatorSetToAPI(valSet entity.ValidatorSet) *api.ValidatorSet {
	return &api.ValidatorSet{
		Version:            valSet.Version,
		RequiredKeyTag:     uint8(valSet.RequiredKeyTag),
		Epoch:              valSet.Epoch,
		CaptureTimestamp:   time.Unix(int64(valSet.CaptureTimestamp), 0).UTC(),
		QuorumThreshold:    valSet.QuorumThreshold.String(),
		PreviousHeaderHash: valSet.PreviousHeaderHash.Hex(),
		Validators: lo.Map(valSet.Validators, func(v entity.Validator, _ int) api.Validator {
			return api.Validator{
				Operator:    v.Operator.Hex(),
				VotingPower: v.VotingPower.String(),
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k entity.Key, _ int) api.Key {
					return api.Key{
						Tag:     uint8(k.Tag),
						Payload: k.Payload,
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v entity.ValidatorVault, _ int) api.ValidatorVault {
					return api.ValidatorVault{
						ChainId:     v.ChainID,
						Vault:       v.Vault.Hex(),
						VotingPower: v.VotingPower.String(),
					}
				}),
			}
		}),
	}
}
