package badger

import (
	"context"
	"encoding/binary"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

// keySignature returns key for a specific signature
// Format: "signature:" + requestID.Bytes() + ":" + validatorIndex (4 bytes BigEndian)
// Epoch is embedded in requestID (first 8 bytes) to enable efficient querying by epoch range
func keySignature(requestID common.Hash, validatorIndex uint32) []byte {
	key := append([]byte("signature:"), requestID.Bytes()...)
	key = append(key, ':')
	// Append validatorIndex as 4-byte BigEndian uint32
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, validatorIndex)
	return append(key, indexBytes...)
}

// keySignatureByEpochPrefix returns prefix for all signatures of a specific epoch
func keySignatureByEpochPrefix(epoch symbiotic.Epoch) []byte {
	return append([]byte("signature:"), epoch.Bytes()...)
}

// keySignaturePrefix returns prefix for all signatures
func keySignaturePrefix() []byte {
	return []byte("signature:")
}

// keySignatureByRequestIDPrefix returns prefix for all signatures of a specific request id within an epoch
func keySignatureByRequestIDPrefix(requestID common.Hash) []byte {
	return append([]byte("signature:"), requestID.Bytes()...)
}

func (r *Repository) saveSignature(
	ctx context.Context,
	validatorIndex uint32,
	sig symbiotic.Signature,
) error {
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

func (r *Repository) GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	return signatures, r.doViewInTx(ctx, "GetAllSignatures", func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keySignatureByRequestIDPrefix(requestID)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
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

func (r *Repository) GetSignaturesByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error) {
	var signatures []symbiotic.Signature

	return signatures, r.doViewInTx(ctx, "GetSignaturesByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Use general prefix for all signatures and start from specific epoch
		startKey := keySignatureByEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keySignaturePrefix()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.Valid(); it.Next() {
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
