package bbolt

import (
	"context"

	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/entity"
)

func (r *Repository) SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error {
	return r.doUpdate(ctx, "SaveNextValsetData", func(tx *bolt.Tx) error {
		txCtx := withTx(ctx, tx)

		if err := r.SaveConfig(txCtx, data.PrevNetworkConfig, data.PrevValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSet(txCtx, data.PrevValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		if err := r.SaveConfig(txCtx, data.NextNetworkConfig, data.NextValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.NextValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSet(txCtx, data.NextValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.NextValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSetMetadata(txCtx, data.ValidatorSetMetadata); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set metadata: %w", err)
		}

		if data.SignatureRequest != nil {
			err := r.SaveSignatureRequest(txCtx, data.ValidatorSetMetadata.RequestID, *data.SignatureRequest)
			if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
				return errors.Errorf("failed to save signature request: %w", err)
			}
		}

		if err := r.saveProofCommitPending(txCtx, data.NextValidatorSet.Epoch, data.ValidatorSetMetadata.RequestID); err != nil {
			return err
		}

		return nil
	})
}
