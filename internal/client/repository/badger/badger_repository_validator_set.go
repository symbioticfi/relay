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

func keyValidatorSetHeader(epoch uint64) []byte {
	return []byte(fmt.Sprintf("validator_set_header:%d", epoch))
}

func keyValidatorByOperator(epoch uint64, operator common.Address) []byte {
	return []byte(fmt.Sprintf("validator:%d:%s", epoch, operator.Hex()))
}

func keyValidatorKeyLookup(epoch uint64, keyTag entity.KeyTag, publicKeyHash common.Hash) []byte {
	return []byte(fmt.Sprintf("validator_key_lookup:%d:%d:%s", epoch, keyTag, publicKeyHash.Hex()))
}

func (r *Repository) SaveValidatorSet(_ context.Context, valset entity.ValidatorSet) error {
	header, err := valset.GetHeader()
	if err != nil {
		return errors.Errorf("failed to create validator set header: %w", err)
	}

	headerBytes, err := validatorSetHeaderToBytes(header)
	if err != nil {
		return errors.Errorf("failed to marshal validator set header: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		// Check if this epoch already exists by checking the header
		headerKey := keyValidatorSetHeader(valset.Epoch)
		_, err := txn.Get(headerKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get validator set header: %w", err)
		}
		if err == nil {
			return errors.Errorf("validator set for epoch %d already exists: %w", valset.Epoch, entity.ErrEntityAlreadyExist)
		}

		// Save the validator set header for its epoch
		err = txn.Set(headerKey, headerBytes)
		if err != nil {
			return errors.Errorf("failed to store validator set header: %w", err)
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

		// Save individual validators and their key indexes
		for _, validator := range valset.Validators {
			validatorBytes, err := validatorToBytes(validator)
			if err != nil {
				return errors.Errorf("failed to marshal validator: %w", err)
			}

			// Save the validator data once
			validatorKey := keyValidatorByOperator(valset.Epoch, validator.Operator)
			err = txn.Set(validatorKey, validatorBytes)
			if err != nil {
				return errors.Errorf("failed to store validator: %w", err)
			}

			// Create an index for each key that points to the validator's operator address
			for _, key := range validator.Keys {
				publicKeyHash := crypto.Keccak256Hash(key.Payload)
				keyLookup := keyValidatorKeyLookup(valset.Epoch, key.Tag, publicKeyHash)
				err = txn.Set(keyLookup, validator.Operator.Bytes())
				if err != nil {
					return errors.Errorf("failed to store validator key lookup: %w", err)
				}
			}
		}

		return nil
	})
}

func (r *Repository) GetValidatorSetHeaderByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSetHeader, error) {
	var header entity.ValidatorSetHeader

	return header, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyValidatorSetHeader(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set header found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err = bytesToValidatorSetHeader(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		return nil
	})
}

func (r *Repository) getAllValidatorsByEpoch(txn *badger.Txn, epoch uint64) (entity.Validators, error) {
	prefix := []byte(fmt.Sprintf("validator:%d:", epoch))
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	var validators entity.Validators
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		value, err := item.ValueCopy(nil)
		if err != nil {
			return nil, errors.Errorf("failed to copy validator value: %w", err)
		}

		validator, err := bytesToValidator(value)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal validator: %w", err)
		}

		validators = append(validators, validator)
	}

	validators.SortByOperatorAddressAsc()

	return validators, nil
}

func (r *Repository) GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		// Get the validator set header
		headerItem, err := txn.Get(keyValidatorSetHeader(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		headerValue, err := headerItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err := bytesToValidatorSetHeader(headerValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		// Get all validators for this epoch
		validators, err := r.getAllValidatorsByEpoch(txn, epoch)
		if err != nil {
			return errors.Errorf("failed to get validators for epoch %d: %w", epoch, err)
		}

		// Build the validator set from header + validators
		vs = entity.ValidatorSet{
			Version:            header.Version,
			RequiredKeyTag:     header.RequiredKeyTag,
			Epoch:              header.Epoch,
			CaptureTimestamp:   header.CaptureTimestamp,
			QuorumThreshold:    header.QuorumThreshold,
			PreviousHeaderHash: header.PreviousHeaderHash,
			Validators:         validators,
			Status:             entity.HeaderCommitted, // Default status for stored validator sets
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSetHeader(_ context.Context) (entity.ValidatorSetHeader, error) {
	var header entity.ValidatorSetHeader

	return header, r.db.View(func(txn *badger.Txn) error {
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

		// Get the validator set header for that epoch in the same transaction
		headerItem, err := txn.Get(keyValidatorSetHeader(latestEpoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set header found for epoch %d: %w", latestEpoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		headerValue, err := headerItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err = bytesToValidatorSetHeader(headerValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSetEpoch(_ context.Context) (uint64, error) {
	var epoch uint64

	return epoch, r.db.View(func(txn *badger.Txn) error {
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

		epoch = binary.BigEndian.Uint64(value)
		return nil
	})
}

func (r *Repository) GetValidatorByKey(_ context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, error) {
	var validator entity.Validator

	publicKeyHash := crypto.Keccak256Hash(publicKey)
	keyLookup := keyValidatorKeyLookup(epoch, keyTag, publicKeyHash)

	return validator, r.db.View(func(txn *badger.Txn) error {
		// First, find the operator address from the key lookup table
		item, err := txn.Get(keyLookup)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator found for epoch %d, keyTag %d, publicKey %x: %w", epoch, keyTag, publicKey, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator key lookup: %w", err)
		}

		operatorBytes, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy operator address value: %w", err)
		}
		operator := common.BytesToAddress(operatorBytes)

		// Now, retrieve the full validator data
		validatorKey := keyValidatorByOperator(epoch, operator)
		item, err = txn.Get(validatorKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				// This would indicate data inconsistency
				return errors.Errorf("found validator key lookup but no validator data for operator %s: %w", operator.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator data: %w", err)
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

	vaults := make([]entity.ValidatorVault, 0, len(dto.Vaults))
	for _, v := range dto.Vaults {
		votingPowerVault, parseOk := new(big.Int).SetString(v.VotingPower, 10)
		if !parseOk {
			return entity.Validator{}, errors.Errorf("failed to parse vault voting power for operator %s: %s", dto.Operator, v.VotingPower)
		}
		vaults = append(vaults, entity.ValidatorVault{
			ChainID:     v.ChainID,
			Vault:       common.HexToAddress(v.Vault),
			VotingPower: entity.ToVotingPower(votingPowerVault),
		})
	}

	return entity.Validator{
		Operator:    operator,
		VotingPower: entity.ToVotingPower(votingPower),
		IsActive:    dto.IsActive,
		Keys:        keys,
		Vaults:      vaults,
	}, nil
}

func validatorSetHeaderToBytes(header entity.ValidatorSetHeader) ([]byte, error) {
	dto := validatorSetHeaderDTO{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              header.Epoch,
		CaptureTimestamp:   header.CaptureTimestamp,
		QuorumThreshold:    header.QuorumThreshold.String(),
		PreviousHeaderHash: header.PreviousHeaderHash.Hex(),
		ValidatorsSszMRoot: header.ValidatorsSszMRoot.Hex(),
	}

	return json.Marshal(dto)
}

func bytesToValidatorSetHeader(data []byte) (entity.ValidatorSetHeader, error) {
	var dto validatorSetHeaderDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to unmarshal validator set header: %w", err)
	}

	quorumThreshold, ok := new(big.Int).SetString(dto.QuorumThreshold, 10)
	if !ok {
		return entity.ValidatorSetHeader{}, errors.Errorf("failed to parse quorum threshold: %s", dto.QuorumThreshold)
	}

	return entity.ValidatorSetHeader{
		Version:            dto.Version,
		RequiredKeyTag:     entity.KeyTag(dto.RequiredKeyTag),
		Epoch:              dto.Epoch,
		CaptureTimestamp:   dto.CaptureTimestamp,
		QuorumThreshold:    entity.ToVotingPower(quorumThreshold),
		PreviousHeaderHash: common.HexToHash(dto.PreviousHeaderHash),
		ValidatorsSszMRoot: common.HexToHash(dto.ValidatorsSszMRoot),
	}, nil
}
