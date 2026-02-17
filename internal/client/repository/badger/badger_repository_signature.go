package badger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func keySignature(requestID common.Hash, validatorIndex uint32) []byte {
	return fmt.Appendf(nil, "signature:%s:%010d", requestID.Hex(), validatorIndex)
}

func keySignatureRequestIDPrefix(requestID common.Hash) []byte {
	return []byte("signature:" + requestID.Hex() + ":")
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

func (r *Repository) GetSignaturesStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	return signatures, r.doViewInTx(ctx, "GetSignaturesStartingFromEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		startKey := keyRequestIDEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keyRequestIDEpochAll()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.Valid(); it.Next() {
			signaturesFromItem, err := r.getSignaturesByEpochFromItem(txn, it)
			if err != nil {
				if errors.Is(err, errCorruptedRequestIDEpochLink) {
					slog.ErrorContext(ctx, errCorruptedRequestIDEpochLink.Error(), "key", string(it.Item().Key()))
					continue
				}
				return err
			}

			signatures = append(signatures, signaturesFromItem...)
		}

		return nil
	})
}

func (r *Repository) GetSignaturesByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	return signatures, r.doViewInTx(ctx, "GetSignaturesByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		startKey := keyRequestIDEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keyRequestIDEpochAll()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.ValidForPrefix(startKey); it.Next() {
			signaturesFromItem, err := r.getSignaturesByEpochFromItem(txn, it)
			if err != nil {
				if errors.Is(err, errCorruptedRequestIDEpochLink) {
					slog.ErrorContext(ctx, errCorruptedRequestIDEpochLink.Error(), "key", string(it.Item().Key()))
					continue
				}
				return err
			}

			signatures = append(signatures, signaturesFromItem...)
		}

		return nil
	})
}

func (r *Repository) getSignaturesByEpochFromItem(txn *badger.Txn, it *badger.Iterator) ([]symbiotic.Signature, error) {
	key := it.Item().Key()

	requestID, err := extractRequestIDFromEpochKey(key)
	if err != nil {
		return nil, errors.Join(errCorruptedRequestIDEpochLink, err)
	}

	return gatAllSignatures(txn, requestID)
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

func signatureToBytes(sig symbiotic.Signature) ([]byte, error) {
	return marshalProto(&pb.Signature{
		MessageHash:  sig.MessageHash,
		KeyTag:       uint32(sig.KeyTag),
		Epoch:        uint64(sig.Epoch),
		Signature:    sig.Signature,
		RawPublicKey: sig.PublicKey.Raw(),
	})
}

func bytesToSignature(value []byte) (symbiotic.Signature, error) {
	signaturePB := &pb.Signature{}
	if err := unmarshalProto(value, signaturePB); err != nil {
		return symbiotic.Signature{}, errors.Errorf("failed to unmarshal signature: %w", err)
	}

	signature := symbiotic.Signature{
		MessageHash: signaturePB.GetMessageHash(),
		KeyTag:      symbiotic.KeyTag(signaturePB.GetKeyTag()),
		Epoch:       symbiotic.Epoch(signaturePB.GetEpoch()),
		Signature:   signaturePB.GetSignature(),
	}

	publicKey, err := crypto.NewPublicKey(signature.KeyTag.Type(), signaturePB.GetRawPublicKey())
	if err != nil {
		return symbiotic.Signature{}, errors.Errorf("failed to get public key from raw: %w", err)
	}

	signature.PublicKey = publicKey

	return signature, nil
}
