package bbolt

import (
	"context"

	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	requestID := aggregationProof.RequestID()

	data, err := codec.AggregationProofToBytes(aggregationProof)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.doUpdate(ctx, "SaveProof", func(tx *bolt.Tx) error {
		if err := putAggregationProofTx(tx, requestID.Bytes(), data, aggregationProof.Epoch); err != nil {
			return err
		}

		// Remove from pending in the same transaction
		pendingKey := epochHashKey(uint64(aggregationProof.Epoch), requestID.Bytes())
		if err := tx.Bucket(bucketAggProofPending).Delete(pendingKey); err != nil {
			return errors.Errorf("failed to delete pending agg proof: %w", err)
		}

		return nil
	})
}
