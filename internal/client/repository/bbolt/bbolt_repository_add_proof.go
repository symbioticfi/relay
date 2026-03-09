package bbolt

import (
	"context"

	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	requestID := aggregationProof.RequestID()

	data, err := codec.AggregationProofToBytes(aggregationProof)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.doUpdate(ctx, "SaveProof", func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAggregationProofs)
		if b.Get(requestID.Bytes()) != nil {
			return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		if err := b.Put(requestID.Bytes(), data); err != nil {
			return errors.Errorf("failed to store aggregation proof: %w", err)
		}

		// Maintain request_id_epochs index
		epochKey := epochHashKey(uint64(aggregationProof.Epoch), requestID.Bytes())
		if tx.Bucket(bucketRequestIDEpochs).Get(epochKey) == nil {
			if err := tx.Bucket(bucketRequestIDEpochs).Put(epochKey, []byte{}); err != nil {
				return errors.Errorf("failed to store request id epoch link: %w", err)
			}
		}

		// Remove from pending in the same transaction
		pendingKey := epochHashKey(uint64(aggregationProof.Epoch), requestID.Bytes())
		tx.Bucket(bucketAggProofPending).Delete(pendingKey) //nolint:errcheck // bbolt Delete only errors on readonly tx

		return nil
	})
}
