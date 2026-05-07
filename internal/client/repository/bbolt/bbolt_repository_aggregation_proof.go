package bbolt

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func putAggregationProofTx(tx *bolt.Tx, requestIDBytes []byte, data []byte, epoch symbiotic.Epoch) error {
	b := tx.Bucket(bucketAggregationProofs)
	if b.Get(requestIDBytes) != nil {
		return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
	}

	if err := b.Put(requestIDBytes, data); err != nil {
		return errors.Errorf("failed to store aggregation proof: %w", err)
	}

	// Maintain request_id_epochs index
	epochKey := epochHashKey(uint64(epoch), requestIDBytes)
	if tx.Bucket(bucketRequestIDEpochs).Get(epochKey) == nil {
		if err := tx.Bucket(bucketRequestIDEpochs).Put(epochKey, []byte{}); err != nil {
			return errors.Errorf("failed to store request id epoch link: %w", err)
		}
	}

	return nil
}

func (r *Repository) saveAggregationProof(ctx context.Context, requestID common.Hash, ap symbiotic.AggregationProof) error {
	data, err := codec.AggregationProofToBytes(ap)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.doBatch(ctx, "saveAggregationProof", func(tx *bolt.Tx) error {
		return putAggregationProofTx(tx, requestID.Bytes(), data, ap.Epoch)
	})
}

func (r *Repository) GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error) {
	var ap symbiotic.AggregationProof

	err := r.doView(ctx, "GetAggregationProof", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketAggregationProofs).Get(requestID.Bytes())
		if v == nil {
			return errors.Errorf("no aggregation proof found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
		}

		var err error
		ap, err = codec.BytesToAggregationProof(v)
		if err != nil {
			return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
		}
		return nil
	})
	return ap, err
}

// GetAggregationProofsByEpoch returns one page of aggregation proofs for the
// given epoch, paginated via opaque cursor `from` (32-byte requestID raw).
// Iterates the request_id_epoch index sorted by requestID within epoch.
// nextFrom == nil signals the last page.
func (r *Repository) GetAggregationProofsByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]symbiotic.AggregationProof, []byte, error) {
	fromHash, err := repoutil.DecodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		proofs []symbiotic.AggregationProof
		lastID common.Hash
	)

	err = r.doView(ctx, "GetAggregationProofsByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketRequestIDEpochs).Cursor()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = epochHashKey(uint64(epoch), fromHash.Bytes())
		}

		k, _ := c.Seek(seekKey)
		if fromHash != (common.Hash{}) && k != nil && bytes.Equal(k, seekKey) {
			k, _ = c.Next()
		}

		for ; k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if pageSize > 0 && len(proofs) >= pageSize {
				return nil
			}
			if len(k) < 40 {
				continue
			}
			id := common.BytesToHash(k[8:40])
			v := tx.Bucket(bucketAggregationProofs).Get(id.Bytes())
			if v == nil {
				continue
			}
			proof, err := codec.BytesToAggregationProof(v)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal aggregation proof", "requestId", id.Hex())
				continue
			}
			proofs = append(proofs, proof)
			lastID = id
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if pageSize == 0 || len(proofs) < pageSize {
		return proofs, nil, nil
	}
	return proofs, repoutil.EncodeHashCursor(lastID), nil
}

func (r *Repository) RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doBatch(ctx, "RemoveAggregationProofPending", func(tx *bolt.Tx) error {
		key := epochHashKey(uint64(epoch), requestID.Bytes())
		b := tx.Bucket(bucketAggProofPending)
		if b.Get(key) == nil {
			return errors.Errorf("pending aggregation proof not found for epoch %d and request id %s: %w", epoch, requestID.Hex(), entity.ErrEntityNotFound)
		}
		return b.Delete(key)
	})
}

func (r *Repository) GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error) {
	var requests []symbiotic.SignatureRequestWithID

	err := r.doView(ctx, "GetSignatureRequestsWithoutAggregationProof", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketAggProofPending).Cursor()

		seekKey := prefix
		if lastHash != (common.Hash{}) {
			seekKey = epochHashKey(uint64(epoch), lastHash.Bytes())
		}

		count := 0
		k, _ := c.Seek(seekKey)
		if lastHash != (common.Hash{}) && k != nil && bytes.Equal(k, seekKey) {
			k, _ = c.Next()
		}

		for ; k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if limit > 0 && count >= limit {
				break
			}
			if len(k) < 40 {
				continue
			}

			requestID := common.BytesToHash(k[8:40])

			// Get the actual signature request
			sigReqKey := epochHashKey(uint64(epoch), requestID.Bytes())
			v := tx.Bucket(bucketSignatureRequests).Get(sigReqKey)
			if v == nil {
				continue
			}

			req, err := codec.BytesToSignatureRequest(v)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}

			requests = append(requests, symbiotic.SignatureRequestWithID{
				SignatureRequest: req,
				RequestID:        requestID,
			})
			count++
		}
		return nil
	})
	return requests, err
}
