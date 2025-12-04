package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error {
	locks := []lock{
		{lockMap: &r.valsetMutexMap, key: data.NextValidatorSet.Epoch},
	}
	if data.PrevValidatorSet.Epoch != data.NextValidatorSet.Epoch {
		locks = append(locks, lock{lockMap: &r.valsetMutexMap, key: data.PrevValidatorSet.Epoch})
	}

	return r.doUpdateInTxWithLock(ctx, "SaveNextValsetData", func(ctx context.Context) error {
		// Save previous validator set and config
		if err := r.saveConfig(ctx, data.PrevNetworkConfig, data.PrevValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSet(ctx, data.PrevValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.PrevValidatorSet.Epoch, err)
		}

		// save bidirectional mapping between next epoch and validator set header request ID
		err := r.saveValsetHeaderRequestIDEpochMapping(ctx, data.NextValidatorSet.Epoch, data.ValidatorSetMetadata.RequestID)
		if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save valset header request ID epoch mapping: %w", err)
		}

		// Save next validator set and config
		if err := r.saveConfig(ctx, data.NextNetworkConfig, data.NextValidatorSet.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save network config for epoch %d: %w", data.NextValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSet(ctx, data.NextValidatorSet); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set for epoch %d: %w", data.NextValidatorSet.Epoch, err)
		}

		if err := r.saveValidatorSetMetadata(ctx, data.ValidatorSetMetadata); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save validator set metadata: %w", err)
		}

		if data.SignatureRequest != nil {
			err := r.SaveSignatureRequest(ctx, data.ValidatorSetMetadata.RequestID, *data.SignatureRequest)
			if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
				return errors.Errorf("failed to save signature request: %w", err)
			}
		}

		// save pending proof commit here
		// we store pending commit request for all nodes and not just current commiters because
		// if committers of this epoch fail then commiters for next epoch should still try to commit old proofs
		if err := r.saveProofCommitPending(ctx, data.NextValidatorSet.Epoch, data.ValidatorSetMetadata.RequestID); err != nil {
			return err
		}

		return nil
	}, locks...)
}

// saveValsetHeaderRequestIDEpochMapping saves bidirectional mapping between epoch and validator set header request ID
func (r *Repository) saveValsetHeaderRequestIDEpochMapping(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTx(ctx, "saveValsetHeaderRequestIDEpochMapping", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Check if request ID → epoch mapping already exists
		requestIDToEpochKey := keyValsetHeaderRequestIDToEpoch(requestID)
		existingEpoch, err := txn.Get(requestIDToEpochKey)
		if err == nil {
			value, err := existingEpoch.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to read existing request ID mapping: %w", err)
			}

			currentEpoch, err := symbiotic.EpochFromBytes(value)
			if err != nil {
				return errors.Errorf("failed to decode existing request ID mapping epoch: %w", err)
			}

			if currentEpoch != epoch {
				return errors.Errorf("valset header request ID already mapped to epoch %d", currentEpoch)
			}

			return nil
		}
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to check existing request ID mapping: %w", err)
		}

		// Check if epoch → request ID mapping already exists
		epochToRequestIDKey := keyValsetHeaderEpochToRequestID(epoch)
		existingRequestID, err := txn.Get(epochToRequestIDKey)
		if err == nil {
			value, err := existingRequestID.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to read existing epoch mapping: %w", err)
			}

			currentRequestID := common.BytesToHash(value)
			if currentRequestID != requestID {
				return errors.Errorf("epoch %d already mapped to different request ID %s", epoch, currentRequestID.Hex())
			}

			return nil
		}

		if !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to check existing epoch mapping: %w", err)
		}

		// Save request ID → epoch mapping
		if err := txn.Set(requestIDToEpochKey, epoch.Bytes()); err != nil {
			return errors.Errorf("failed to save request ID to epoch mapping: %w", err)
		}

		// Save epoch → request ID mapping
		if err := txn.Set(epochToRequestIDKey, requestID.Bytes()); err != nil {
			return errors.Errorf("failed to save epoch to request ID mapping: %w", err)
		}

		return nil
	})
}
