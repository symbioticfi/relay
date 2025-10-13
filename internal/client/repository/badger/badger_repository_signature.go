package badger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func keySignature(requestID common.Hash, validatorIndex uint32) []byte {
	return []byte(fmt.Sprintf("signature:%s:%010d", requestID.Hex(), validatorIndex))
}

// keySignaturePrefix returns prefix for all signatures of a request id
func keySignaturePrefix(requestID common.Hash) []byte {
	return []byte("signature:" + requestID.Hex() + ":")
}

func (r *Repository) saveSignature(ctx context.Context, validatorIndex uint32, sig symbiotic.SignatureExtended) error {
	txn := getTxn(ctx)
	if txn == nil {
		return errors.New("no transaction found in context, use signature processor in order to store signatures")
	}

	bytes, err := signatureToBytes(sig)
	if err != nil {
		return errors.Errorf("failed to marshal signature: %w", err)
	}

	key := keySignature(sig.RequestID(), validatorIndex)
	err = txn.Set(key, bytes)
	if err != nil {
		return errors.Errorf("failed to store signature: %w", err)
	}
	return nil
}

func (r *Repository) GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.SignatureExtended, error) {
	var signatures []symbiotic.SignatureExtended

	return signatures, r.doViewInTx(ctx, "GetAllSignatures", func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keySignaturePrefix(requestID)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature value: %w", err)
			}

			sig, err := bytesToSignature(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature: %w", err)
			}

			signatures = append(signatures, sig)
		}

		return nil
	})
}

func (r *Repository) GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.SignatureExtended, error) {
	var signature symbiotic.SignatureExtended

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

func signatureToBytes(sig symbiotic.SignatureExtended) ([]byte, error) {
	return marshalAndCompress(&v1.Signature{
		MessageHash: sig.MessageHash,
		KeyTag:      uint32(sig.KeyTag),
		Epoch:       uint64(sig.Epoch),
		Signature:   sig.Signature,
		PublicKey:   sig.PublicKey,
	})
}

func bytesToSignature(value []byte) (symbiotic.SignatureExtended, error) {
	pb := &v1.Signature{}
	if err := unmarshalAndDecompress(value, pb); err != nil {
		return symbiotic.SignatureExtended{}, errors.Errorf("failed to unmarshal signature: %w", err)
	}

	return symbiotic.SignatureExtended{
		MessageHash: pb.GetMessageHash(),
		KeyTag:      symbiotic.KeyTag(pb.GetKeyTag()),
		Epoch:       symbiotic.Epoch(pb.GetEpoch()),
		PublicKey:   pb.GetPublicKey(),
		Signature:   pb.GetSignature(),
	}, nil
}
