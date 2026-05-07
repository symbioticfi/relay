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
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
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

// GetSignaturesByEpoch returns one page of signatures for the given epoch,
// paginated via 36-byte composite cursor `from` (requestID || BE32(vIdx)).
// Pagination is per-signature so pageSize is honored exactly even when groups
// are large. nextFrom == nil signals the last page.
func (r *Repository) GetSignaturesByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]symbiotic.Signature, []byte, error) {
	fromRequestID, fromVIdx, err := repoutil.DecodeSignatureCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		signatures    []symbiotic.Signature
		lastRequestID common.Hash
		lastVIdx      uint32
	)

	err = r.doView(ctx, "GetSignaturesByEpoch", func(tx *bolt.Tx) error {
		idxPrefix := epochBytes(uint64(epoch))
		idxC := tx.Bucket(bucketRequestIDEpochs).Cursor()

		idxSeekKey := idxPrefix
		if fromRequestID != (common.Hash{}) {
			idxSeekKey = epochHashKey(uint64(epoch), fromRequestID.Bytes())
		}

	outer:
		for k, _ := idxC.Seek(idxSeekKey); k != nil && bytes.HasPrefix(k, idxPrefix); k, _ = idxC.Next() {
			if len(k) < 40 {
				continue
			}
			currentRequestID := common.BytesToHash(k[8:40])

			// fromRequestID is zero only when from == nil, so a hash match
			// here implies a real cursor — fromVIdx may legitimately be 0.
			startVIdx := uint32(0)
			if currentRequestID == fromRequestID {
				startVIdx = fromVIdx + 1
			}

			sigC := tx.Bucket(bucketSignatures).Cursor()
			sigPrefix := currentRequestID.Bytes()
			sigSeekKey := signatureKey(currentRequestID.Bytes(), startVIdx)

			for sk, sv := sigC.Seek(sigSeekKey); sk != nil && bytes.HasPrefix(sk, sigPrefix); sk, sv = sigC.Next() {
				if pageSize > 0 && len(signatures) >= pageSize {
					break outer
				}
				if len(sk) != 36 {
					continue
				}
				vIdx := binary.BigEndian.Uint32(sk[32:])
				sig, err := codec.BytesToSignature(sv)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to unmarshal signature", "key", hex.EncodeToString(sk), "error", err)
					continue
				}
				signatures = append(signatures, sig)
				lastRequestID = currentRequestID
				lastVIdx = vIdx
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if pageSize == 0 || len(signatures) < pageSize {
		return signatures, nil, nil
	}
	return signatures, repoutil.EncodeSignatureCursor(lastRequestID, lastVIdx), nil
}

func (r *Repository) GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	if raw, ok := r.signatureMapCache.Load(requestID); ok {
		sm := raw.(entity.SignatureMap)
		return sm.Clone(), nil
	}

	sm, err := r.loadSignatureMap(ctx, requestID)
	if err != nil {
		return entity.SignatureMap{}, err
	}
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

	return r.doBatch(ctx, "SaveSignatureRequest", func(tx *bolt.Tx) error {
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

// GetSignatureRequestsWithIDByEpoch returns one page of signature requests for
// the given epoch, paginated via opaque cursor `from` (32-byte requestID raw).
// nextFrom == nil signals the last page.
func (r *Repository) GetSignatureRequestsWithIDByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]entity.SignatureRequestWithID, []byte, error) {
	fromHash, err := repoutil.DecodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		requests []entity.SignatureRequestWithID
		lastID   common.Hash
	)

	err = r.doView(ctx, "GetSignatureRequestsWithIDByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketSignatureRequests).Cursor()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = epochHashKey(uint64(epoch), fromHash.Bytes())
		}

		k, v := c.Seek(seekKey)
		if fromHash != (common.Hash{}) && k != nil && bytes.Equal(k, seekKey) {
			k, v = c.Next()
		}

		for ; k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if pageSize > 0 && len(requests) >= pageSize {
				return nil
			}
			if len(k) < 40 {
				continue
			}
			id := common.BytesToHash(k[8:40])
			req, err := codec.BytesToSignatureRequest(v)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}
			requests = append(requests, entity.SignatureRequestWithID{
				RequestID:        id,
				SignatureRequest: req,
			})
			lastID = id
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if pageSize == 0 || len(requests) < pageSize {
		return requests, nil, nil
	}
	return requests, repoutil.EncodeHashCursor(lastID), nil
}

// GetSignatureRequestIDsByEpoch returns one page of request IDs for the given
// epoch (keys only — no value materialization). Cursor format same as
// GetSignatureRequestsWithIDByEpoch.
func (r *Repository) GetSignatureRequestIDsByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]common.Hash, []byte, error) {
	fromHash, err := repoutil.DecodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		ids    []common.Hash
		lastID common.Hash
	)

	err = r.doView(ctx, "GetSignatureRequestIDsByEpoch", func(tx *bolt.Tx) error {
		prefix := epochBytes(uint64(epoch))
		c := tx.Bucket(bucketSignatureRequests).Cursor()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = epochHashKey(uint64(epoch), fromHash.Bytes())
		}

		k, _ := c.Seek(seekKey)
		if fromHash != (common.Hash{}) && k != nil && bytes.Equal(k, seekKey) {
			k, _ = c.Next()
		}

		for ; k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if pageSize > 0 && len(ids) >= pageSize {
				return nil
			}
			if len(k) < 40 {
				continue
			}
			id := common.BytesToHash(k[8:40])
			ids = append(ids, id)
			lastID = id
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if pageSize == 0 || len(ids) < pageSize {
		return ids, nil, nil
	}
	return ids, repoutil.EncodeHashCursor(lastID), nil
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
	return r.doBatch(ctx, "RemoveSignaturePending", func(tx *bolt.Tx) error {
		key := epochHashKey(uint64(epoch), requestID.Bytes())
		return tx.Bucket(bucketSignaturePending).Delete(key)
	})
}
