package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
)

const (
	latestValidatorSetKey = "latest_validator_set"
)

func keyValidatorSet(epoch uint64) []byte {
	return []byte(fmt.Sprintf("validator_set:%d", epoch))
}

func keyValidator(epoch uint64, keyTag entity.KeyTag, publicKeyHash common.Hash) []byte {
	return []byte(fmt.Sprintf("validator:%d:%d:%s", epoch, keyTag, publicKeyHash.Hex()))
}

func (r *Repository) SaveValidatorSet(_ context.Context, valset entity.ValidatorSet) error {
	bytes, err := validatorSetToBytes(valset)
	if err != nil {
		return errors.Errorf("failed to marshal validator set: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		// Save by epoch
		epochKey := keyValidatorSet(valset.Epoch)
		_, err := txn.Get(epochKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get validator set: %w", err)
		}
		if err == nil {
			return errors.Errorf("validator set for epoch %d already exists: %w", valset.Epoch, entity.ErrEntityAlreadyExist)
		}

		// Save the validator set for its epoch
		err = txn.Set(epochKey, bytes)
		if err != nil {
			return errors.Errorf("failed to store validator set: %w", err)
		}

		// Check if this is a newer epoch than the latest one
		latestItem, err := txn.Get([]byte(latestValidatorSetKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get latest validator set: %w", err)
		}

		shouldUpdateLatest := true
		if err == nil {
			latestValue, err := latestItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy latest validator set value: %w", err)
			}
			latestVs, err := bytesToValidatorSet(latestValue)
			if err != nil {
				return errors.Errorf("failed to unmarshal latest validator set: %w", err)
			}
			shouldUpdateLatest = latestVs.Epoch < valset.Epoch
		}

		// Update latest validator set only if this is a newer epoch
		if shouldUpdateLatest {
			err = txn.Set([]byte(latestValidatorSetKey), bytes)
			if err != nil {
				return errors.Errorf("failed to store latest validator set: %w", err)
			}
		}

		// Save individual validator indexes
		for _, validator := range valset.Validators {
			validatorBytes, err := validatorToBytes(validator)
			if err != nil {
				return errors.Errorf("failed to marshal validator: %w", err)
			}

			// Save validator for each key tag it has
			for _, key := range validator.Keys {
				publicKeyHash := crypto.Keccak256Hash(key.Payload)
				validatorKey := keyValidator(valset.Epoch, key.Tag, publicKeyHash)

				err = txn.Set(validatorKey, validatorBytes)
				if err != nil {
					return errors.Errorf("failed to store validator index: %w", err)
				}
			}
		}

		return nil
	})
}

func (r *Repository) GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyValidatorSet(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set value: %w", err)
		}

		vs, err = bytesToValidatorSet(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSet(_ context.Context) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(latestValidatorSetKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set value: %w", err)
		}

		vs, err = bytesToValidatorSet(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal latest validator set: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetValidatorByKey(_ context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, error) {
	var validator entity.Validator

	publicKeyHash := crypto.Keccak256Hash(publicKey)
	key := keyValidator(epoch, keyTag, publicKeyHash)

	return validator, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator found for epoch %d, keyTag %d, publicKey %x: %w", epoch, keyTag, publicKey, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator value: %w", err)
		}

		validator, err = bytesToValidator(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator: %w", err)
		}

		return nil
	})
}

func validatorToBytes(validator entity.Validator) ([]byte, error) {
	dto := validatorDTO{
		Operator:    validator.Operator.Hex(),
		VotingPower: validator.VotingPower.String(),
		IsActive:    validator.IsActive,
		Keys: lo.Map(validator.Keys, func(k entity.ValidatorKey, _ int) keyDTO {
			return keyDTO{
				Tag:     uint8(k.Tag),
				Payload: k.Payload,
			}
		}),
		Vaults: lo.Map(validator.Vaults, func(v entity.ValidatorVault, _ int) validatorVaultDTO {
			return validatorVaultDTO{
				ChainID:     v.ChainID,
				Vault:       v.Vault.Hex(),
				VotingPower: v.VotingPower.String(),
			}
		}),
	}

	return json.Marshal(dto)
}

func bytesToValidator(data []byte) (entity.Validator, error) {
	var dto validatorDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.Validator{}, errors.Errorf("failed to unmarshal validator: %w", err)
	}

	operator := common.HexToAddress(dto.Operator)

	votingPower, ok := new(big.Int).SetString(dto.VotingPower, 10)
	if !ok {
		return entity.Validator{}, errors.Errorf("failed to parse voting power: %s", dto.VotingPower)
	}

	keys := lo.Map(dto.Keys, func(k keyDTO, _ int) entity.ValidatorKey {
		return entity.ValidatorKey{
			Tag:     entity.KeyTag(k.Tag),
			Payload: k.Payload,
		}
	})

	vaults := lo.Map(dto.Vaults, func(v validatorVaultDTO, _ int) entity.ValidatorVault {
		votingPowerVault, parseOk := new(big.Int).SetString(v.VotingPower, 10)
		if !parseOk {
			return entity.ValidatorVault{}
		}
		return entity.ValidatorVault{
			ChainID:     v.ChainID,
			Vault:       common.HexToAddress(v.Vault),
			VotingPower: entity.ToVotingPower(votingPowerVault),
		}
	})

	return entity.Validator{
		Operator:    operator,
		VotingPower: entity.ToVotingPower(votingPower),
		IsActive:    dto.IsActive,
		Keys:        keys,
		Vaults:      vaults,
	}, nil
}
