package badger

import (
	"context"
	"encoding/binary"
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
	latestValidatorSetEpochKey = "latest_validator_set_epoch"
)

func keyValidatorSet(epoch uint64) []byte {
	return []byte(fmt.Sprintf("validator_set:%d", epoch))
}

func keyValidatorSetMeta(epoch uint64) []byte {
	return []byte(fmt.Sprintf("validator_set_meta:%d", epoch))
}

func keyValidator(epoch uint64, keyTag entity.KeyTag, publicKeyHash common.Hash) []byte {
	return []byte(fmt.Sprintf("validator:%d:%d:%s", epoch, keyTag, publicKeyHash.Hex()))
}

func (r *Repository) SaveValidatorSet(_ context.Context, valset entity.ValidatorSet) error {
	valsetMeta, err := valset.MakeMeta()
	if err != nil {
		return errors.Errorf("failed to create validator set meta: %w", err)
	}
	bytes, err := validatorSetToBytes(valset)
	if err != nil {
		return errors.Errorf("failed to marshal validator set: %w", err)
	}

	metaBytes, err := validatorSetMetaToBytes(valsetMeta)
	if err != nil {
		return errors.Errorf("failed to marshal validator set meta: %w", err)
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

		// Save the validator set meta for its epoch
		metaKey := keyValidatorSetMeta(valset.Epoch)
		err = txn.Set(metaKey, metaBytes)
		if err != nil {
			return errors.Errorf("failed to store validator set meta: %w", err)
		}

		// Check if this is a newer epoch than the latest one
		latestItem, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		shouldUpdateLatest := true
		if err == nil {
			latestValue, err := latestItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
			}
			latestEpoch := binary.BigEndian.Uint64(latestValue)
			shouldUpdateLatest = latestEpoch < valset.Epoch
		}

		// Update latest validator set epoch only if this is a newer epoch
		if shouldUpdateLatest {
			epochBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(epochBytes, valset.Epoch)
			err = txn.Set([]byte(latestValidatorSetEpochKey), epochBytes)
			if err != nil {
				return errors.Errorf("failed to store latest validator set epoch: %w", err)
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

func (r *Repository) GetValidatorSetMetaByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSetMeta, error) {
	var meta entity.ValidatorSetMeta

	return meta, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyValidatorSetMeta(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set meta found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set meta: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set meta value: %w", err)
		}

		meta, err = bytesToValidatorSetMeta(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set meta: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		var err error
		vs, err = r.getValidatorSetByEpochTx(txn, epoch)
		return err
	})
}

func (r *Repository) getValidatorSetByEpochTx(txn *badger.Txn, epoch uint64) (entity.ValidatorSet, error) {
	item, err := txn.Get(keyValidatorSet(epoch))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return entity.ValidatorSet{}, errors.Errorf("no validator set found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}
		return entity.ValidatorSet{}, errors.Errorf("failed to get validator set: %w", err)
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to copy validator set value: %w", err)
	}

	vs, err := bytesToValidatorSet(value)
	if err != nil {
		return entity.ValidatorSet{}, errors.Errorf("failed to unmarshal validator set: %w", err)
	}

	return vs, nil
}

func (r *Repository) GetLatestValidatorSetMeta(_ context.Context) (entity.ValidatorSetMeta, error) {
	var meta entity.ValidatorSetMeta

	return meta, r.db.View(func(txn *badger.Txn) error {
		// Get the latest epoch
		item, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
		}

		latestEpoch := binary.BigEndian.Uint64(value)

		// Get the validator set meta for that epoch in the same transaction
		metaItem, err := txn.Get(keyValidatorSetMeta(latestEpoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set meta found for epoch %d: %w", latestEpoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set meta: %w", err)
		}

		metaValue, err := metaItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set meta value: %w", err)
		}

		meta, err = bytesToValidatorSetMeta(metaValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set meta: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSet(_ context.Context) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		// Get the latest epoch
		item, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
		}

		latestEpoch := binary.BigEndian.Uint64(value)

		// Get the validator set for that epoch in the same transaction
		vs, err = r.getValidatorSetByEpochTx(txn, latestEpoch)
		return err
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

func validatorSetMetaToBytes(meta entity.ValidatorSetMeta) ([]byte, error) {
	dto := validatorSetMetaDTO{
		Version:            meta.Version,
		RequiredKeyTag:     uint8(meta.RequiredKeyTag),
		Epoch:              meta.Epoch,
		CaptureTimestamp:   meta.CaptureTimestamp,
		QuorumThreshold:    meta.QuorumThreshold.String(),
		PreviousHeaderHash: meta.PreviousHeaderHash.Hex(),
		TotalVotingPower:   meta.TotalActiveVotingPower.String(),
		ValidatorsSszMRoot: meta.ValidatorsSszMRoot.Hex(),
	}

	return json.Marshal(dto)
}

func bytesToValidatorSetMeta(data []byte) (entity.ValidatorSetMeta, error) {
	var dto validatorSetMetaDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.ValidatorSetMeta{}, errors.Errorf("failed to unmarshal validator set meta: %w", err)
	}

	quorumThreshold, ok := new(big.Int).SetString(dto.QuorumThreshold, 10)
	if !ok {
		return entity.ValidatorSetMeta{}, errors.Errorf("failed to parse quorum threshold: %s", dto.QuorumThreshold)
	}

	totalVotingPower, ok := new(big.Int).SetString(dto.TotalVotingPower, 10)
	if !ok {
		return entity.ValidatorSetMeta{}, errors.Errorf("failed to parse total voting power: %s", dto.TotalVotingPower)
	}

	return entity.ValidatorSetMeta{
		Version:                dto.Version,
		RequiredKeyTag:         entity.KeyTag(dto.RequiredKeyTag),
		Epoch:                  dto.Epoch,
		CaptureTimestamp:       dto.CaptureTimestamp,
		QuorumThreshold:        entity.ToVotingPower(quorumThreshold),
		PreviousHeaderHash:     common.HexToHash(dto.PreviousHeaderHash),
		TotalActiveVotingPower: entity.ToVotingPower(totalVotingPower),
		ValidatorsSszMRoot:     common.HexToHash(dto.ValidatorsSszMRoot),
	}, nil
}
