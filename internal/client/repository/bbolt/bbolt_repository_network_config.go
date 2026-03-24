package bbolt

import (
	"context"

	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	data, err := codec.NetworkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	return r.doUpdate(ctx, "SaveConfig", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		b := tx.Bucket(bucketNetworkConfigs)
		if b.Get(ek) != nil {
			return errors.Errorf("network config for epoch %d already exists: %w", epoch, entity.ErrEntityAlreadyExist)
		}
		return b.Put(ek, data)
	})
}

func (r *Repository) GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	var config symbiotic.NetworkConfig

	err := r.doView(ctx, "GetConfigByEpoch", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketNetworkConfigs).Get(epochBytes(uint64(epoch)))
		if v == nil {
			return errors.Errorf("no network config found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}

		var err error
		config, err = codec.BytesToNetworkConfig(v)
		return err
	})
	return config, err
}
