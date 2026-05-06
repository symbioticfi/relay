package badger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func keySignature(requestID common.Hash, validatorIndex uint32) []byte {
	return fmt.Appendf(nil, "signature:%s:%010d", requestID.Hex(), validatorIndex)
}

func keySignatureRequestIDPrefix(requestID common.Hash) []byte {
	return []byte("signature:" + requestID.Hex() + ":")
}

// extractValidatorIndexFromSignatureKey parses the trailing 10-digit vIdx
// out of `signature:<hex>:<NNNNNNNNNN>`. Returns an error if the key shape is
// unexpected (corrupted store).
func extractValidatorIndexFromSignatureKey(key []byte) (uint32, error) {
	idx := bytes.LastIndexByte(key, ':')
	if idx < 0 || idx == len(key)-1 {
		return 0, errors.Errorf("invalid signature key shape: %q", string(key))
	}
	v, err := strconv.ParseUint(string(key[idx+1:]), 10, 32)
	if err != nil {
		return 0, errors.Errorf("invalid validator index in key %q: %w", string(key), err)
	}
	return uint32(v), nil
}

func (r *Repository) saveSignature(
	ctx context.Context,
	validatorIndex uint32,
	sig symbiotic.Signature,
) error {
	bytes, err := signatureToBytes(sig)
	if err != nil {
		return errors.Errorf("failed to marshal signature: %w", err)
	}

	return r.doUpdateInTx(ctx, "saveSignature", func(ctx context.Context) error {
		requestID := sig.RequestID()
		valueKey := keySignature(requestID, validatorIndex)

		txn := getTxn(ctx)
		if err = txn.Set(valueKey, bytes); err != nil {
			return errors.Errorf("failed to store signature: %w", err)
		}

		reqIDEpochKey := keyRequestIDEpoch(sig.Epoch, requestID)

		_, err = txn.Get(reqIDEpochKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get request id epoch link: %w", err)
		}
		if err == nil {
			return nil
		}

		if err = txn.Set(reqIDEpochKey, []byte{}); err != nil {
			return errors.Errorf("failed to store request id epoch link: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	return signatures, r.doViewInTx(ctx, "GetAllSignatures", func(ctx context.Context) error {
		var err error
		signatures, err = gatAllSignatures(getTxn(ctx), requestID)
		return err
	})
}

func (r *Repository) GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error) {
	var signature symbiotic.Signature

	err := r.doViewInTx(ctx, "GetSignatureByIndex", func(ctx context.Context) error {
		txn := getTxn(ctx)
		key := keySignature(requestID, validatorIndex)

		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return entity.ErrEntityNotFound
			}
			return errors.Errorf("failed to get signature: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy signature value: %w", err)
		}

		sig, err := bytesToSignature(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal signature: %w", err)
		}

		signature = sig
		return nil
	})

	return signature, err
}

// GetSignaturesByEpoch returns one page of signatures for the given epoch,
// paginated via 36-byte composite cursor `from` (requestID || BE32(vIdx)).
// Pagination is per-signature so pageSize is honored exactly even when groups
// (signatures sharing a requestID) are large. nextFrom == nil signals last page.
func (r *Repository) GetSignaturesByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]symbiotic.Signature, []byte, error) {
	fromRequestID, fromVIdx, err := decodeSignatureCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		signatures    []symbiotic.Signature
		lastRequestID common.Hash
		lastVIdx      uint32
		filledFull    bool
	)

	err = r.doViewInTx(ctx, "GetSignaturesByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		idxPrefix := keyRequestIDEpochPrefix(epoch)
		idxOpts := badger.DefaultIteratorOptions
		idxOpts.Prefix = keyRequestIDEpochAll()
		idxOpts.PrefetchValues = false
		idxIt := txn.NewIterator(idxOpts)
		defer idxIt.Close()

		// Seek to the index entry for fromRequestID (or start). We do NOT
		// skip-on-exact-match: when the cursor lands on a known requestID we
		// still need to read its remaining signatures (vIdx > fromVIdx).
		idxSeekKey := idxPrefix
		if fromRequestID != (common.Hash{}) {
			idxSeekKey = keyRequestIDEpoch(epoch, fromRequestID)
		}

		count := 0
	outer:
		for idxIt.Seek(idxSeekKey); idxIt.ValidForPrefix(idxPrefix); idxIt.Next() {
			currentRequestID, err := extractRequestIDFromEpochKey(idxIt.Item().Key())
			if err != nil {
				slog.ErrorContext(ctx, errCorruptedRequestIDEpochLink.Error(), "key", string(idxIt.Item().Key()))
				continue
			}

			// Within the cursor's group, skip already-returned signatures.
			startVIdx := uint32(0)
			if currentRequestID == fromRequestID && fromVIdx > 0 {
				startVIdx = fromVIdx + 1
			}

			sigPrefix := keySignatureRequestIDPrefix(currentRequestID)
			sigOpts := badger.DefaultIteratorOptions
			sigOpts.Prefix = sigPrefix
			sigIt := txn.NewIterator(sigOpts)

			for sigIt.Seek(keySignature(currentRequestID, startVIdx)); sigIt.ValidForPrefix(sigPrefix); sigIt.Next() {
				if pageSize > 0 && count >= pageSize {
					sigIt.Close()
					filledFull = true
					break outer
				}
				vIdx, err := extractValidatorIndexFromSignatureKey(sigIt.Item().Key())
				if err != nil {
					sigIt.Close()
					return err
				}
				value, err := sigIt.Item().ValueCopy(nil)
				if err != nil {
					sigIt.Close()
					return errors.Errorf("failed to copy signature value: %w", err)
				}
				sig, err := bytesToSignature(value)
				if err != nil {
					sigIt.Close()
					return errors.Errorf("failed to unmarshal signature: %w", err)
				}
				signatures = append(signatures, sig)
				lastRequestID = currentRequestID
				lastVIdx = vIdx
				count++
			}
			sigIt.Close()
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !filledFull {
		return signatures, nil, nil
	}
	return signatures, encodeSignatureCursor(lastRequestID, lastVIdx), nil
}

func gatAllSignatures(txn *badger.Txn, requestID common.Hash) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	prefix := keySignatureRequestIDPrefix(requestID)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix

	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		value, err := item.ValueCopy(nil)
		if err != nil {
			return nil, errors.Errorf("failed to copy signature value: %w", err)
		}

		sig, err := bytesToSignature(value)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal signature: %w", err)
		}

		signatures = append(signatures, sig)
	}

	return signatures, nil
}

var (
	signatureToBytes = codec.SignatureToBytes
	bytesToSignature = codec.BytesToSignature
)
