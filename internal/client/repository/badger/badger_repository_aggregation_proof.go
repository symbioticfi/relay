package badger

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// keyAggregationProof returns key for a specific aggregation proof
// Format: "aggregation_proof:" + requestID bytes (where first 8 bytes of requestID is epoch)
// Epoch is embedded in requestID to enable efficient querying by epoch range
func keyAggregationProof(requestID common.Hash) []byte {
	return append([]byte("aggregation_proof:"), requestID.Bytes()...)
}

// keyAggregationProofByEpochPrefix returns prefix for all aggregation proofs of a specific epoch
func keyAggregationProofByEpochPrefix(epoch symbiotic.Epoch) []byte {
	return append([]byte("aggregation_proof:"), epoch.Bytes()...)
}

// keyAggregationProofPrefix returns prefix for all aggregation proofs
func keyAggregationProofPrefix() []byte {
	return []byte("aggregation_proof:")
}

func keyAggregationProofPending(epoch symbiotic.Epoch, requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("aggregation_proof_pending:%d:%s", epoch, requestID.Hex()))
}

func keyAggregationProofPendingEpochPrefix(epoch symbiotic.Epoch) []byte {
	return []byte(fmt.Sprintf("aggregation_proof_pending:%d:", epoch))
}

func (r *Repository) saveAggregationProof(ctx context.Context, requestID common.Hash, ap symbiotic.AggregationProof) error {
	proofBytes, err := aggregationProofToBytes(ap)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.doUpdateInTxWithLock(ctx, "saveAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)
		key := keyAggregationProof(requestID)
		_, err := txn.Get(key)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(key, proofBytes)
		if err != nil {
			return errors.Errorf("failed to store aggregation proof: %w", err)
		}
		return nil
	}, &r.proofsMutexMap, requestID)
}

func (r *Repository) GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error) {
	var ap symbiotic.AggregationProof

	return ap, r.doViewInTx(ctx, "GetAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)
		key := keyAggregationProof(requestID)
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no aggregation proof found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy aggregation proof value: %w", err)
		}

		ap, err = bytesToAggregationProof(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetAggregationProofsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.AggregationProof, error) {
	var proofs []symbiotic.AggregationProof

	return proofs, r.doViewInTx(ctx, "GetAggregationProofsByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Use general prefix for all aggregation proofs and start from specific epoch
		startKey := keyAggregationProofByEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keyAggregationProofPrefix()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy aggregation proof value: %w", err)
			}

			proof, err := bytesToAggregationProof(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
			}

			proofs = append(proofs, proof)
		}

		return nil
	})
}

func aggregationProofToBytes(ap symbiotic.AggregationProof) ([]byte, error) {
	return marshalProto(&pb.AggregationProof{
		MessageHash: ap.MessageHash,
		KeyTag:      uint32(ap.KeyTag),
		Epoch:       uint64(ap.Epoch),
		Proof:       ap.Proof,
	})
}

func bytesToAggregationProof(value []byte) (symbiotic.AggregationProof, error) {
	aggregationProof := &pb.AggregationProof{}
	if err := unmarshalProto(value, aggregationProof); err != nil {
		return symbiotic.AggregationProof{}, errors.Errorf("failed to unmarshal aggregation proof: %w", err)
	}

	return symbiotic.AggregationProof{
		MessageHash: aggregationProof.GetMessageHash(),
		KeyTag:      symbiotic.KeyTag(aggregationProof.GetKeyTag()),
		Epoch:       symbiotic.Epoch(aggregationProof.GetEpoch()),
		Proof:       aggregationProof.GetProof(),
	}, nil
}

func (r *Repository) saveAggregationProofPending(ctx context.Context, requestID common.Hash, epoch symbiotic.Epoch) error {
	return r.doUpdateInTx(ctx, "saveAggregationProofPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofPending(epoch, requestID)

		_, err := txn.Get(pendingKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to check pending aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("pending aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		// Store just a marker (empty value) - we don't need the full request data here
		err = txn.Set(pendingKey, []byte{})
		if err != nil {
			return errors.Errorf("failed to store pending aggregation proof: %w", err)
		}
		return nil
	})
}

func (r *Repository) RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTx(ctx, "RemoveAggregationProofPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofPending(epoch, requestID)

		// Check if exists before removing
		_, err := txn.Get(pendingKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("pending aggregation proof not found for epoch %d and request id %s: %w", epoch, requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to check pending aggregation proof: %w", err)
		}

		err = txn.Delete(pendingKey)
		if err != nil {
			return errors.Errorf("failed to delete pending aggregation proof: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error) {
	var requests []symbiotic.SignatureRequestWithID

	return requests, r.doViewInTx(ctx, "GetSignatureRequestsWithoutAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Iterate through pending aggregation proof markers
		prefix := keyAggregationProofPendingEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false // We don't need the values, just the keys
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if lastHash != (common.Hash{}) {
			// Subsequent pages: seek to the record after lastHash
			seekKey = keyAggregationProofPending(epoch, lastHash)
		}

		count := 0
		it.Seek(seekKey)
		// If we're seeking from a specific hash and positioned exactly on that key, skip it (already returned in previous page)
		if lastHash != (common.Hash{}) && it.ValidForPrefix(prefix) && bytes.Equal(it.Item().Key(), seekKey) {
			it.Next()
		}

		for ; it.ValidForPrefix(prefix); it.Next() {
			// Stop if we've reached the limit
			if limit > 0 && count >= limit {
				break
			}

			// Extract request id from the pending key: "aggregation_proof_pending:epoch:request_id"
			item := it.Item()
			key := string(item.Key())

			// Find the hash part after the second colon
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				return errors.Errorf("invalid pending aggregation proof key format: %s", key)
			}

			requestIDStr := parts[2]
			requestID := common.HexToHash(requestIDStr)

			// Get the actual signature request
			sigReqKey := keySignatureRequest(epoch, requestID)
			sigReqItem, err := txn.Get(sigReqKey)
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					// This shouldn't happen - pending marker exists but signature request doesn't
					// Skip this entry and continue
					continue
				}
				return errors.Errorf("failed to get signature request for hash %s: %w", requestIDStr, err)
			}

			value, err := sigReqItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature request value: %w", err)
			}

			req, err := bytesToSignatureRequest(value)
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
}
