package bbolt

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	err := r.doView(ctx, "GetAllSignatures", func(tx *bolt.Tx) error {
		var err error
		signatures, err = getAllSignatures(tx, requestID)
		return err
	})
	return signatures, err
}

func getAllSignatures(tx *bolt.Tx, requestID common.Hash) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	prefix := requestID.Bytes() // 32 bytes
	c := tx.Bucket(bucketSignatures).Cursor()

	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		sig, err := codec.BytesToSignature(v)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal signature: %w", err)
		}
		signatures = append(signatures, sig)
	}

	return signatures, nil
}

func (r *Repository) GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error) {
	var sig symbiotic.Signature

	err := r.doView(ctx, "GetSignatureByIndex", func(tx *bolt.Tx) error {
		key := signatureKey(requestID.Bytes(), validatorIndex)
		v := tx.Bucket(bucketSignatures).Get(key)
		if v == nil {
			return entity.ErrEntityNotFound
		}

		var err error
		sig, err = codec.BytesToSignature(v)
		if err != nil {
			return errors.Errorf("failed to unmarshal signature: %w", err)
		}
		return nil
	})

	return sig, err
}

func (r *Repository) GetSignaturesStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	err := r.doView(ctx, "GetSignaturesStartingFromEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketRequestIDEpochs).Cursor()

		for k, _ := c.Seek(prefix); k != nil; k, _ = c.Next() {
			if len(k) < 40 {
				continue
			}
			requestID := common.BytesToHash(k[8:40])
			sigs, err := getAllSignatures(tx, requestID)
			if err != nil {
				slog.ErrorContext(ctx, "Corrupted request id epoch link", "key", hex.EncodeToString(k))
				continue
			}
			signatures = append(signatures, sigs...)
		}
		return nil
	})
	return signatures, err
}

func (r *Repository) GetSignaturesByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	err := r.doView(ctx, "GetSignaturesByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketRequestIDEpochs).Cursor()

		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if len(k) < 40 {
				continue
			}
			requestID := common.BytesToHash(k[8:40])
			sigs, err := getAllSignatures(tx, requestID)
			if err != nil {
				slog.ErrorContext(ctx, "Corrupted request id epoch link", "key", hex.EncodeToString(k))
				continue
			}
			signatures = append(signatures, sigs...)
		}
		return nil
	})
	return signatures, err
}

func (r *Repository) GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	if raw, ok := r.signatureMapCache.Load(requestID); ok {
		sm := raw.(entity.SignatureMap)
		return sm.Clone(), nil
	}

	res, err, _ := r.signatureMapLoader.Do(requestID.Hex(), func() (any, error) {
		return r.rebuildSignatureMap(ctx, requestID)
	})
	if err != nil {
		return entity.SignatureMap{}, err
	}

	rebuilt := res.(entity.SignatureMap)
	actual, _ := r.signatureMapCache.LoadOrStore(requestID, rebuilt)
	sm := actual.(entity.SignatureMap)
	return sm.Clone(), nil
}

func (r *Repository) rebuildSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	var sm entity.SignatureMap

	err := r.doView(ctx, "rebuildSignatureMap", func(tx *bolt.Tx) error {
		// Scan stored signatures: extract activeIndices and discover epoch from first signature
		sigPrefix := requestID.Bytes()
		sc := tx.Bucket(bucketSignatures).Cursor()

		var epoch symbiotic.Epoch
		var activeIndices []uint32
		epochFound := false

		for k, v := sc.Seek(sigPrefix); k != nil && bytes.HasPrefix(k, sigPrefix); k, v = sc.Next() {
			if len(k) != 36 { // 32 (requestID) + 4 (activeIndex)
				continue
			}
			activeIndices = append(activeIndices, binary.BigEndian.Uint32(k[32:]))

			if !epochFound {
				sig, err := codec.BytesToSignature(v)
				if err != nil {
					return errors.Errorf("failed to unmarshal signature: %w", err)
				}
				epoch = sig.Epoch
				epochFound = true
			}
		}

		if !epochFound {
			return errors.Errorf("no signatures found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
		}

		countVal := tx.Bucket(bucketActiveValCounts).Get(epochBytes(uint64(epoch)))
		if countVal == nil {
			return errors.Errorf("no active validator count for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}
		totalActive := binary.BigEndian.Uint32(countVal)

		// Build activeIndex → votingPower from stored validators
		votingPowers := make(map[uint32]symbiotic.VotingPower)
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketValidators).Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			val, storedActiveIdx, err := codec.BytesToValidator(v)
			if err != nil {
				return errors.Errorf("failed to unmarshal validator: %w", err)
			}
			if val.IsActive {
				votingPowers[storedActiveIdx] = val.VotingPower
			}
		}

		sm = entity.NewSignatureMap(requestID, epoch, totalActive)
		for _, activeIdx := range activeIndices {
			vp, ok := votingPowers[activeIdx]
			if !ok {
				continue
			}
			_ = sm.SetValidatorPresent(activeIdx, vp)
		}

		return nil
	})

	if err != nil {
		return entity.SignatureMap{}, err
	}

	return sm, nil
}

func (r *Repository) SaveSignatureRequest(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error {
	data, err := codec.SignatureRequestToBytes(req)
	if err != nil {
		return errors.Errorf("failed to marshal signature request: %w", err)
	}
	primaryKey := epochHashKey(uint64(req.RequiredEpoch), requestID.Bytes())
	pendingKey := epochHashKey(uint64(req.RequiredEpoch), requestID.Bytes())
	epochData := epochBytes(uint64(req.RequiredEpoch))

	return r.doUpdate(ctx, "SaveSignatureRequest", func(tx *bolt.Tx) error {
		// Save signature request
		b := tx.Bucket(bucketSignatureRequests)
		if b.Get(primaryKey) != nil {
			return errors.Errorf("signature request already exists: %w", entity.ErrEntityAlreadyExist)
		}

		if err := b.Put(primaryKey, data); err != nil {
			return errors.Errorf("failed to store signature request: %w", err)
		}

		// Save request ID index: requestID → epoch bytes
		if err := tx.Bucket(bucketRequestIDIndex).Put(requestID.Bytes(), epochData); err != nil {
			return errors.Errorf("failed to store request id index: %w", err)
		}

		// Save pending signature marker
		pendingBucket := tx.Bucket(bucketSignaturePending)
		if pendingBucket.Get(pendingKey) != nil {
			return nil // Already pending
		}
		if err := pendingBucket.Put(pendingKey, []byte{}); err != nil {
			return errors.Errorf("failed to store pending signature: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error) {
	var req symbiotic.SignatureRequest

	err := r.doView(ctx, "GetSignatureRequest", func(tx *bolt.Tx) error {
		// Look up epoch from index
		epochVal := tx.Bucket(bucketRequestIDIndex).Get(requestID.Bytes())
		if epochVal == nil {
			return errors.Errorf("no signature request found for request id %s: %w", requestID.String(), entity.ErrEntityNotFound)
		}

		epoch := binary.BigEndian.Uint64(epochVal)
		key := epochHashKey(epoch, requestID.Bytes())
		v := tx.Bucket(bucketSignatureRequests).Get(key)
		if v == nil {
			return errors.Errorf("failed to get signature request: %w", entity.ErrEntityNotFound)
		}

		var err error
		req, err = codec.BytesToSignatureRequest(v)
		if err != nil {
			return errors.Errorf("failed to unmarshal signature request: %w", err)
		}
		return nil
	})
	return req, err
}

func (r *Repository) GetSignatureRequestsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequest, error) {
	var requests []symbiotic.SignatureRequest

	err := r.doView(ctx, "GetSignatureRequestsByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		b := tx.Bucket(bucketSignatureRequests)
		c := b.Cursor()

		seekKey := prefix
		if lastHash != (common.Hash{}) {
			seekKey = epochHashKey(uint64(epoch), lastHash.Bytes())
		}

		count := 0
		k, v := c.Seek(seekKey)
		if lastHash != (common.Hash{}) && k != nil && bytes.Equal(k, seekKey) {
			k, v = c.Next()
		}

		for ; k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if limit > 0 && count >= limit {
				break
			}
			req, err := codec.BytesToSignatureRequest(v)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}
			requests = append(requests, req)
			count++
		}
		return nil
	})
	return requests, err
}

func (r *Repository) GetSignatureRequestsWithIDByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]entity.SignatureRequestWithID, error) {
	var requests []entity.SignatureRequestWithID

	err := r.doView(ctx, "GetSignatureRequestsWithIDByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketSignatureRequests).Cursor()

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if len(k) < 40 {
				continue
			}
			requestID := common.BytesToHash(k[8:40])
			req, err := codec.BytesToSignatureRequest(v)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}
			requests = append(requests, entity.SignatureRequestWithID{
				RequestID:        requestID,
				SignatureRequest: req,
			})
		}
		return nil
	})
	return requests, err
}

func (r *Repository) GetSignatureRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]common.Hash, error) {
	var ids []common.Hash

	err := r.doView(ctx, "GetSignatureRequestIDsByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketSignatureRequests).Cursor()

		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if len(k) < 40 {
				continue
			}
			ids = append(ids, common.BytesToHash(k[8:40]))
		}
		return nil
	})
	return ids, err
}

func (r *Repository) GetSignaturePending(ctx context.Context, limit int) ([]common.Hash, error) {
	var requests []common.Hash

	err := r.doView(ctx, "GetSignaturePending", func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketSignaturePending).Cursor()
		count := 0

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if limit > 0 && count >= limit {
				break
			}
			if len(k) < 40 {
				continue
			}
			requests = append(requests, common.BytesToHash(k[8:40]))
			count++
		}
		return nil
	})
	return requests, err
}

func (r *Repository) RemoveSignaturePending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdate(ctx, "RemoveSignaturePending", func(tx *bolt.Tx) error {
		key := epochHashKey(uint64(epoch), requestID.Bytes())
		return tx.Bucket(bucketSignaturePending).Delete(key)
	})
}
