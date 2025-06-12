package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

func (r *Repository) SaveConfig(_ context.Context, config entity.NetworkConfig, epoch uint64) error {
	configBytes, err := networkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	// Store in Badger
	err = r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(keyNetworkConfig(epoch))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get network config: %w", err)
		}
		if err == nil {
			return errors.Errorf("network config for epoch %d already exists: %w", epoch, entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyNetworkConfig(epoch), configBytes)
		if err != nil {
			return errors.Errorf("failed to store network config: %w", err)
		}
		return nil
	})

	if err != nil {
		return errors.Errorf("failed to save network config for epoch %d: %w", epoch, err)
	}

	return nil
}

func (r *Repository) GetConfigByEpoch(_ context.Context, epoch uint64) (entity.NetworkConfig, error) {
	var config entity.NetworkConfig

	return config, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyNetworkConfig(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no network config found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get network config: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy network config value: %w", err)
		}

		config, err = bytesToNetworkConfig(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal network config: %w", err)
		}

		return nil
	})
}
