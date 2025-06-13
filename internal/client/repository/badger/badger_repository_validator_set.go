package badger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

const (
	latestValidatorSetKey = "latest_validator_set"
)

func keyValidatorSet(epoch uint64) []byte {
	return []byte(fmt.Sprintf("validator_set:%d", epoch))
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
