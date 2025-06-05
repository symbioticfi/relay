package entity

import (
	"math/big"
	"sort"

	"github.com/samber/lo"
)

type ValidatorSetExtra struct {
	Version              uint8
	RequiredKeyTag       KeyTag
	Config               Config
	DomainEip712         Eip712Domain
	Subnetwork           []byte
	Keys                 []OperatorWithKeys
	OperatorVotingPowers []OperatorVotingPower
	Epoch                *big.Int
	CaptureTimestamp     *big.Int
}

func (v ValidatorSetExtra) MakeValidatorSet() ValidatorSet {
	// Create validators map to consolidate voting powers and keys
	validatorsMap := make(map[string]*Validator)

	// Process voting powers
	for _, vp := range v.OperatorVotingPowers {
		operatorAddr := vp.Operator.Hex()
		if _, exists := validatorsMap[operatorAddr]; !exists {
			validatorsMap[operatorAddr] = &Validator{
				Operator:    vp.Operator,
				VotingPower: big.NewInt(0),
				IsActive:    false, // Default to active, will filter later
				Keys:        []Key{},
				Vaults:      []Vault{},
			}
		}

		// Add vaults and their voting powers
		for _, vault := range vp.Vaults {
			validatorsMap[operatorAddr].VotingPower = new(big.Int).Add(
				validatorsMap[operatorAddr].VotingPower,
				vault.VotingPower,
			)

			// Add vault to validator's vaults
			validatorsMap[operatorAddr].Vaults = append(validatorsMap[operatorAddr].Vaults, Vault{
				Vault:       vault.Vault,
				VotingPower: vault.VotingPower,
			})
		}

		// Sort vaults by address in ascending order
		sort.Slice(validatorsMap[operatorAddr].Vaults, func(i, j int) bool {
			// Compare voting powers (higher first)
			return validatorsMap[operatorAddr].Vaults[i].Vault.Cmp(validatorsMap[operatorAddr].Vaults[j].Vault) < 0
		})
	}

	// Process required keys
	for _, rk := range v.Keys { // TODO: get required key tags from validator set config and fill with nils if needed
		operatorAddr := rk.Operator.Hex()
		if validator, exists := validatorsMap[operatorAddr]; exists {
			// Add all keys for this operator
			for _, key := range rk.Keys {
				validator.Keys = append(validator.Keys, Key{
					Tag:     key.Tag,
					Payload: key.Payload,
				})
			}
		}
	}

	validators := lo.Map(lo.Values(validatorsMap), func(item *Validator, _ int) Validator {
		return *item
	})
	// Sort validators by voting power in descending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (higher first)
		return validators[i].VotingPower.Cmp(validators[j].VotingPower) > 0
	})

	totalActive := 0

	totalActiveVotingPower := big.NewInt(0)
	for i := range validators {
		// Check minimum voting power if configured
		if validators[i].VotingPower.Cmp(v.Config.MinInclusionVotingPower) < 0 {
			break
		}

		// Check if validator has at least one key
		if len(validators[i].Keys) == 0 {
			continue
		}

		totalActive++
		validators[i].IsActive = true

		if v.Config.MaxVotingPower.Int64() != 0 {
			if validators[i].VotingPower.Cmp(v.Config.MaxVotingPower) > 0 {
				validators[i].VotingPower = new(big.Int).Set(v.Config.MaxVotingPower)
			}
		}
		// Add to total active voting power if validator is active
		totalActiveVotingPower = new(big.Int).Add(totalActiveVotingPower, validators[i].VotingPower)

		if v.Config.MaxValidatorsCount.Int64() != 0 {
			if totalActive >= int(v.Config.MaxValidatorsCount.Int64()) {
				break
			}
		}
	}

	// Sort validators by address in ascending order
	sort.Slice(validators, func(i, j int) bool {
		// Compare voting powers (higher first)
		return validators[i].Operator.Cmp(validators[j].Operator) < 0
	})

	// Create the validator set
	return ValidatorSet{
		Version:                v.Version,
		TotalActiveVotingPower: totalActiveVotingPower,
		Validators:             validators,
	}
}
