package aggregator

import (
	"middleware-offchain/core/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

func getActiveValidators(allValidators []entity.Validator) []entity.Validator {
	activeValidators := make([]entity.Validator, 0)
	for _, validator := range allValidators {
		if validator.IsActive {
			activeValidators = append(activeValidators, validator)
		}
	}
	return activeValidators
}

func toValidatorsData(signerValidators []entity.Validator, allValidators []entity.Validator, requiredKeyTag entity.KeyTag) ([]proof.ValidatorData, error) {
	activeValidators := getActiveValidators(allValidators)
	valset := make([]proof.ValidatorData, 0)
	for i := range activeValidators {
		for _, key := range activeValidators[i].Keys {
			if key.Tag == requiredKeyTag {
				g1, err := bls.DeserializeG1(key.Payload)
				if err != nil {
					return nil, errors.Errorf("failed to deserialize G1: %w", err)
				}
				validatorData := proof.ValidatorData{Key: *g1.G1Affine, VotingPower: activeValidators[i].VotingPower.Int, IsNonSigner: true}

				for _, signer := range signerValidators {
					if signer.Operator.Cmp(activeValidators[i].Operator) == 0 {
						validatorData.IsNonSigner = false
						break
					}
				}

				valset = append(valset, validatorData)
				break
			}
		}
	}
	return proof.NormalizeValset(valset), nil
}

func validatorSetMimcAccumulator(valset []entity.Validator, requiredKeyTag entity.KeyTag) (common.Hash, error) {
	validatorsData, err := toValidatorsData([]entity.Validator{}, valset, requiredKeyTag)
	if err != nil {
		return common.Hash{}, err
	}
	return common.Hash(proof.HashValset(validatorsData)), nil
}
