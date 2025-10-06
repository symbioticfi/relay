package badger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func keyAggregationProof(requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("aggregation_proof:%s", requestID.Hex()))
}

func keyAggregationProofPending(epoch entity.Epoch, requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("aggregation_proof_pending:%d:%s", epoch, requestID.Hex()))
}

func keyAggregationProofPendingEpochPrefix(epoch entity.Epoch) []byte {
	return []byte(fmt.Sprintf("aggregation_proof_pending:%d:", epoch))
}

func (r *Repository) saveAggregationProof(ctx context.Context, requestID common.Hash, ap entity.AggregationProof) error {
	proofBytes, err := aggregationProofToBytes(ap)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.doUpdateInTx(ctx, "saveAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)
		_, err := txn.Get(keyAggregationProof(requestID))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyAggregationProof(requestID), proofBytes)
		if err != nil {
			return errors.Errorf("failed to store aggregation proof: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetAggregationProof(ctx context.Context, requestID common.Hash) (entity.AggregationProof, error) {
	var ap entity.AggregationProof

	return ap, r.doViewInTx(ctx, "GetAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)
		item, err := txn.Get(keyAggregationProof(requestID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no aggregation proof found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy network config value: %w", err)
		}

		ap, err = bytesToAggregationProof(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
		}

		return nil
	})
}

type aggregationProofDTO struct {
	MessageHash []byte `json:"message_hash"`
	KeyTag      uint8  `json:"key_tag"`
	Epoch       uint64 `json:"epoch"`
	Proof       []byte `json:"proof"`
}

func aggregationProofToBytes(ap entity.AggregationProof) ([]byte, error) {
	dto := aggregationProofDTO{
		MessageHash: ap.MessageHash,
		KeyTag:      uint8(ap.KeyTag),
		Epoch:       uint64(ap.Epoch),
		Proof:       ap.Proof,
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return data, nil
}

func bytesToAggregationProof(value []byte) (entity.AggregationProof, error) {
	var dto aggregationProofDTO
	if err := json.Unmarshal(value, &dto); err != nil {
		return entity.AggregationProof{}, errors.Errorf("failed to unmarshal aggregation proof: %w", err)
	}

	return entity.AggregationProof{
		MessageHash: dto.MessageHash,
		KeyTag:      entity.KeyTag(dto.KeyTag),
		Epoch:       entity.Epoch(dto.Epoch),
		Proof:       dto.Proof,
	}, nil
}

func (r *Repository) saveAggregationProofPending(ctx context.Context, requestID common.Hash, epoch entity.Epoch) error {
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

func (r *Repository) RemoveAggregationProofPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error {
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

func (r *Repository) GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequestWithID, error) {
	var requests []entity.SignatureRequestWithID

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

			requests = append(requests, entity.SignatureRequestWithID{
				SignatureRequest: req,
				RequestID:        requestID,
			})
			count++
		}

		return nil
	})
}
