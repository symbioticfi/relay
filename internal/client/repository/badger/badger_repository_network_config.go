package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const networkConfigPrefix = "network_config:"

func keyNetworkConfig(epoch symbiotic.Epoch) []byte {
	return epochKey(networkConfigPrefix, epoch)
}

func (r *Repository) SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	configBytes, err := networkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	return r.doUpdateInTx(ctx, "saveConfig", func(ctx context.Context) error {
		txn := getTxn(ctx)
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
}

func (r *Repository) GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	var config symbiotic.NetworkConfig

	return config, r.doViewInTx(ctx, "GetConfigByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
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

var (
	networkConfigToBytes = codec.NetworkConfigToBytes
	bytesToNetworkConfig = codec.BytesToNetworkConfig
)
