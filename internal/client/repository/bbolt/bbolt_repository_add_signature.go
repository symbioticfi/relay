package bbolt

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error {
	start := time.Now()

	if err := r.saveSignatureWithPending(ctx, activeIndex, signature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	saveDuration := time.Since(start)

	cacheStart := time.Now()
	signatureMap, err := r.addToSignatureMapCache(ctx, signature.RequestID(), activeIndex, validator.VotingPower)
	if err != nil {
		return errors.Errorf("failed to update signature map cache: %w", err)
	}

	slog.DebugContext(ctx, "Saved signature for validator",
		"activeIndex", activeIndex,
		"requestId", signature.RequestID().Hex(),
		"epoch", signature.Epoch,
		"totalSignatures", signatureMap.SignedValidatorsBitmap.GetCardinality(),
		"presentValidators", signatureMap.SignedValidatorsBitmap.ToArray(),
		"saveDuration", saveDuration,
		"cacheDuration", time.Since(cacheStart),
	)

	return nil
}

func (r *Repository) saveSignatureWithPending(ctx context.Context, validatorIndex uint32, sig symbiotic.Signature) error {
	data, err := codec.SignatureToBytes(sig)
	if err != nil {
		return errors.Errorf("failed to marshal signature: %w", err)
	}

	requestID := sig.RequestID()
	epochKey := epochHashKey(uint64(sig.Epoch), requestID.Bytes())

	return r.doBatch(ctx, "saveSignatureWithPending", func(tx *bolt.Tx) error {
		key := signatureKey(requestID.Bytes(), validatorIndex)

		if tx.Bucket(bucketSignatures).Get(key) != nil {
			return nil
		}

		if err := tx.Bucket(bucketSignatures).Put(key, data); err != nil {
			return errors.Errorf("failed to store signature: %w", err)
		}

		// Maintain request_id_epochs index
		if tx.Bucket(bucketRequestIDEpochs).Get(epochKey) == nil {
			if err := tx.Bucket(bucketRequestIDEpochs).Put(epochKey, []byte{}); err != nil {
				return errors.Errorf("failed to store request id epoch link: %w", err)
			}
		}

		// Save pending proof if aggregation proof does not exist yet
		if sig.KeyTag.Type().AggregationKey() {
			if tx.Bucket(bucketAggregationProofs).Get(requestID.Bytes()) != nil {
				return nil
			}
		}

		pendingBucket := tx.Bucket(bucketAggProofPending)
		if pendingBucket.Get(epochKey) == nil {
			if err := pendingBucket.Put(epochKey, []byte{}); err != nil {
				return errors.Errorf("failed to store pending proof: %w", err)
			}
		}

		return nil
	})
}

func (r *Repository) addToSignatureMapCache(ctx context.Context, requestID common.Hash, activeIndex uint32, votingPower symbiotic.VotingPower) (entity.SignatureMap, error) {
	for {
		raw, ok := r.signatureMapCache.Load(requestID)
		if !ok {
			base, err := r.loadSignatureMap(ctx, requestID)
			if err != nil {
				return entity.SignatureMap{}, errors.Errorf("failed to load signature map: %w", err)
			}
			actual, loaded := r.signatureMapCache.LoadOrStore(requestID, base)
			if !loaded {
				raw = base
			} else {
				raw = actual
			}
		}

		old := raw.(entity.SignatureMap)
		if old.SignedValidatorsBitmap.Contains(activeIndex) {
			return old, nil
		}

		cloned := old.Clone()
		_ = cloned.SetValidatorPresent(activeIndex, votingPower)
		if r.signatureMapCache.CompareAndSwap(requestID, raw, cloned) {
			return cloned, nil
		}
	}
}

// loadSignatureMap rebuilds the signature map for requestID from the database, deduplicating
// concurrent rebuilds via singleflight. Returns ErrEntityNotFound transparently if the request
// has no signatures yet — the caller's CAS path will populate the map on subsequent calls.
func (r *Repository) loadSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	res, err, _ := r.signatureMapLoader.Do(requestID.Hex(), func() (any, error) {
		sm, err := r.rebuildSignatureMap(ctx, requestID)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return entity.SignatureMap{}, err
		}
		return sm, nil
	})
	if err != nil {
		return entity.SignatureMap{}, err
	}
	return res.(entity.SignatureMap), nil
}
