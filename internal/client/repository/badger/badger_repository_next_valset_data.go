package badger

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
)

func (r *Repository) SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error {
	return r.doUpdateInTx(ctx, "SaveNextValsetData", func(ctx context.Context) error {
		// Save previous validator set and config
		if err := r.SaveConfig(ctx, data.PrevNetworkConfig, data.PrevValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		if err := r.SaveValidatorSet(ctx, data.PrevValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		// Save next validator set and config
		if err := r.SaveConfig(ctx, data.NextNetworkConfig, data.NextValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		if err := r.SaveValidatorSet(ctx, data.NextValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.NextValidatorSet.Epoch, err)
		}

		if data.SignatureRequest != nil {
			err := r.SaveSignatureRequest(ctx, data.ValidatorSetMetadata.RequestID, *data.SignatureRequest)
			if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
				return errors.Errorf("failed to get signature request: %w", err)
			}
		}

		// save pending proof commit here
		// we store pending commit request for all nodes and not just current commiters because
		// if committers of this epoch fail then commiters for next epoch should still try to commit old proofs
		if err := r.SaveProofCommitPending(ctx, data.NextValidatorSet.Epoch, data.ValidatorSetMetadata.RequestID); err != nil {
			return err
		}

		return nil
	})
}
