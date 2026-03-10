package bbolt

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

var (
	metaLatestValidatorSetEpoch     = []byte("latest_validator_set_epoch")
	metaLatestAggregatedValsetEpoch = []byte("latest_aggregated_validator_set_epoch")
	metaFirstUncommittedValsetEpoch = []byte("first_uncommitted_validator_set_epoch")
)

func (r *Repository) saveValidatorSet(ctx context.Context, valset symbiotic.ValidatorSet) error {
	if err := valset.Validators.CheckIsSortedByOperatorAddressAsc(); err != nil {
		return errors.Errorf("validators must be sorted by operator address ascending: %w", err)
	}

	headerBytes, err := codec.ValidatorSetHeaderToBytes(valset)
	if err != nil {
		return errors.Errorf("failed to marshal validator set header: %w", err)
	}

	return r.doUpdate(ctx, "saveValidatorSet", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(valset.Epoch))

		// Check if exists
		if tx.Bucket(bucketValidatorSetHeaders).Get(ek) != nil {
			return errors.Errorf("validator set for epoch %d already exists: %w", valset.Epoch, entity.ErrEntityAlreadyExist)
		}

		if err := tx.Bucket(bucketValidatorSetHeaders).Put(ek, headerBytes); err != nil {
			return errors.Errorf("failed to store validator set header: %w", err)
		}

		if err := tx.Bucket(bucketValidatorSetStatus).Put(ek, []byte{uint8(valset.Status)}); err != nil {
			return errors.Errorf("failed to store validator set status: %w", err)
		}

		// Update latest epoch
		if err := updateLatestEpochIfNeeded(tx, metaLatestValidatorSetEpoch, valset.Epoch); err != nil {
			return err
		}

		activeIndex := uint32(0)
		for _, v := range valset.Validators {
			currentActiveIndex := uint32(0)
			if v.IsActive {
				currentActiveIndex = activeIndex
				activeIndex++
			}

			valBytes, err := codec.ValidatorToBytes(v, currentActiveIndex)
			if err != nil {
				return errors.Errorf("failed to marshal validator: %w", err)
			}

			vk := epochOperatorKey(uint64(valset.Epoch), v.Operator.Bytes())
			if err := tx.Bucket(bucketValidators).Put(vk, valBytes); err != nil {
				return errors.Errorf("failed to store validator: %w", err)
			}

			for _, key := range v.Keys {
				pubKeyHash := crypto.Keccak256Hash(key.Payload)
				lk := validatorKeyLookupKey(uint64(valset.Epoch), uint32(key.Tag), pubKeyHash.Bytes())
				if err := tx.Bucket(bucketValidatorKeyLookups).Put(lk, v.Operator.Bytes()); err != nil {
					return errors.Errorf("failed to store validator key lookup: %w", err)
				}
			}
		}

		countBytes := uint32Bytes(activeIndex)
		if err := tx.Bucket(bucketActiveValCounts).Put(ek, countBytes); err != nil {
			return errors.Errorf("failed to store active validator count: %w", err)
		}

		return nil
	})
}

func updateLatestEpochIfNeeded(tx *bolt.Tx, key []byte, epoch symbiotic.Epoch) error {
	b := tx.Bucket(bucketMeta)
	existing := b.Get(key)
	if existing != nil {
		existingEpoch := symbiotic.Epoch(binary.BigEndian.Uint64(existing))
		if existingEpoch >= epoch {
			return nil
		}
	}
	return b.Put(key, epochBytes(uint64(epoch)))
}

func (r *Repository) SaveFirstUncommittedValidatorSetEpoch(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "SaveFirstUncommittedValidatorSetEpoch", func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).Put(metaFirstUncommittedValsetEpoch, epochBytes(uint64(epoch)))
	})
}

func (r *Repository) UpdateValidatorSetStatusAndRemovePendingProof(ctx context.Context, valset symbiotic.ValidatorSet) error {
	return r.doUpdate(ctx, "UpdateValidatorSetStatusAndRemovePendingProof", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(valset.Epoch))

		if tx.Bucket(bucketValidatorSetStatus).Get(ek) == nil {
			return errors.Errorf("failed to get validator set status key: %w", entity.ErrEntityNotFound)
		}

		if err := tx.Bucket(bucketValidatorSetStatus).Put(ek, []byte{uint8(symbiotic.HeaderCommitted)}); err != nil {
			return errors.Errorf("failed to store validator set status: %w", err)
		}

		// Remove pending proof commit (ignore not found)
		tx.Bucket(bucketAggProofCommits).Delete(ek) //nolint:errcheck // bbolt Delete only errors on readonly tx

		return nil
	})
}

func (r *Repository) UpdateValidatorSetStatus(ctx context.Context, epoch symbiotic.Epoch, status symbiotic.ValidatorSetStatus) error {
	return r.doUpdate(ctx, "UpdateValidatorSetStatus", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		b := tx.Bucket(bucketValidatorSetStatus)

		if b.Get(ek) == nil {
			return errors.Errorf("failed to get validator set status key: %w", entity.ErrEntityNotFound)
		}

		if err := b.Put(ek, []byte{uint8(status)}); err != nil {
			return errors.Errorf("failed to store validator set status: %w", err)
		}

		if status >= symbiotic.HeaderAggregated {
			if err := updateLatestEpochIfNeeded(tx, metaLatestAggregatedValsetEpoch, epoch); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetValidatorSetHeaderByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetHeader, error) {
	var header symbiotic.ValidatorSetHeader

	err := r.doView(ctx, "GetValidatorSetHeaderByEpoch", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketValidatorSetHeaders).Get(epochBytes(uint64(epoch)))
		if v == nil {
			return errors.Errorf("no validator set header found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}

		var err error
		header, err = codec.BytesToValidatorSetHeader(v)
		return err
	})
	return header, err
}

func (r *Repository) getAllValidatorsByEpoch(tx *bolt.Tx, epoch symbiotic.Epoch) (symbiotic.Validators, error) {
	prefix := epochBytes(uint64(epoch))
	c := tx.Bucket(bucketValidators).Cursor()
	var validators symbiotic.Validators

	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		val, _, err := codec.BytesToValidator(v)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal validator: %w", err)
		}
		validators = append(validators, val)
	}

	validators.SortByOperatorAddressAsc()
	return validators, nil
}

func (r *Repository) GetValidatorSetsStartingFromEpoch(ctx context.Context, startEpoch symbiotic.Epoch) ([]symbiotic.ValidatorSet, error) {
	var sets []symbiotic.ValidatorSet

	err := r.doView(ctx, "GetValidatorSetsStartingFromEpoch", func(tx *bolt.Tx) error {
		seekKey := epochBytes(uint64(startEpoch))
		c := tx.Bucket(bucketValidatorSetHeaders).Cursor()

		for k, headerValue := c.Seek(seekKey); k != nil; k, headerValue = c.Next() {
			header, err := codec.BytesToValidatorSetHeader(headerValue)
			if err != nil {
				return errors.Errorf("failed to unmarshal validator set header: %w", err)
			}

			statusVal := tx.Bucket(bucketValidatorSetStatus).Get(k)
			if statusVal == nil || len(statusVal) != 1 {
				continue
			}
			status := symbiotic.ValidatorSetStatus(statusVal[0])

			validators, err := r.getAllValidatorsByEpoch(tx, header.Epoch)
			if err != nil {
				return errors.Errorf("failed to get validators for epoch %d: %w", header.Epoch, err)
			}

			aggIndices, commIndices, err := codec.ExtractAdditionalInfoFromHeaderData(headerValue)
			if err != nil {
				return errors.Errorf("failed to extract bitmap indices: %w", err)
			}

			sets = append(sets, symbiotic.ValidatorSet{
				Version:           header.Version,
				RequiredKeyTag:    header.RequiredKeyTag,
				Epoch:             header.Epoch,
				CaptureTimestamp:  header.CaptureTimestamp,
				QuorumThreshold:   header.QuorumThreshold,
				Validators:        validators,
				Status:            status,
				AggregatorIndices: aggIndices,
				CommitterIndices:  commIndices,
			})
		}
		return nil
	})
	return sets, err
}

func (r *Repository) GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	var vs symbiotic.ValidatorSet

	err := r.doView(ctx, "GetValidatorSetByEpoch", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))

		headerValue := tx.Bucket(bucketValidatorSetHeaders).Get(ek)
		if headerValue == nil {
			return errors.Errorf("no validator set found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}

		header, err := codec.BytesToValidatorSetHeader(headerValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		statusVal := tx.Bucket(bucketValidatorSetStatus).Get(ek)
		if statusVal == nil || len(statusVal) != 1 {
			return errors.Errorf("no validator set status found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}

		validators, err := r.getAllValidatorsByEpoch(tx, epoch)
		if err != nil {
			return errors.Errorf("failed to get validators: %w", err)
		}

		aggIndices, commIndices, err := codec.ExtractAdditionalInfoFromHeaderData(headerValue)
		if err != nil {
			return errors.Errorf("failed to extract bitmap indices: %w", err)
		}

		vs = symbiotic.ValidatorSet{
			Version:           header.Version,
			RequiredKeyTag:    header.RequiredKeyTag,
			Epoch:             header.Epoch,
			CaptureTimestamp:  header.CaptureTimestamp,
			QuorumThreshold:   header.QuorumThreshold,
			Validators:        validators,
			Status:            symbiotic.ValidatorSetStatus(statusVal[0]),
			AggregatorIndices: aggIndices,
			CommitterIndices:  commIndices,
		}
		return nil
	})
	return vs, err
}

func (r *Repository) GetLatestValidatorSetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error) {
	var header symbiotic.ValidatorSetHeader

	err := r.doView(ctx, "GetLatestValidatorSetHeader", func(tx *bolt.Tx) error {
		epochVal := tx.Bucket(bucketMeta).Get(metaLatestValidatorSetEpoch)
		if epochVal == nil {
			return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
		}

		headerVal := tx.Bucket(bucketValidatorSetHeaders).Get(epochVal)
		if headerVal == nil {
			return errors.Errorf("no validator set header found: %w", entity.ErrEntityNotFound)
		}

		var err error
		header, err = codec.BytesToValidatorSetHeader(headerVal)
		return err
	})
	return header, err
}

func (r *Repository) GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	err := r.doView(ctx, "GetLatestValidatorSetEpoch", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketMeta).Get(metaLatestValidatorSetEpoch)
		if v == nil {
			return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
		}
		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(v))
		return nil
	})
	return epoch, err
}

func (r *Repository) GetOldestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	err := r.doView(ctx, "GetOldestValidatorSetEpoch", func(tx *bolt.Tx) error {
		k, _ := tx.Bucket(bucketValidatorSetHeaders).Cursor().First()
		if k == nil {
			return errors.Errorf("no validator set headers found: %w", entity.ErrEntityNotFound)
		}
		if len(k) != 8 {
			return errors.New("invalid epoch key length")
		}
		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(k))
		return nil
	})
	return epoch, err
}

func (r *Repository) GetActiveValidatorCountByEpoch(ctx context.Context, epoch symbiotic.Epoch) (uint32, error) {
	var count uint32

	err := r.doView(ctx, "GetActiveValidatorCountByEpoch", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketActiveValCounts).Get(epochBytes(uint64(epoch)))
		if v == nil {
			return errors.Errorf("no active validator count found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}
		count = binary.BigEndian.Uint32(v)
		return nil
	})
	return count, err
}

func (r *Repository) GetFirstUncommittedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	err := r.doView(ctx, "GetFirstUncommittedValidatorSetEpoch", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketMeta).Get(metaFirstUncommittedValsetEpoch)
		if v == nil {
			return nil // No uncommitted epoch found, return zero
		}
		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(v))
		return nil
	})
	return epoch, err
}

func (r *Repository) GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error) {
	publicKeyHash := crypto.Keccak256Hash(publicKey)
	lk := validatorKeyLookupKey(uint64(epoch), uint32(keyTag), publicKeyHash.Bytes())

	var val symbiotic.Validator
	var activeIndex uint32

	err := r.doView(ctx, "GetValidatorByKey", func(tx *bolt.Tx) error {
		operatorBytes := tx.Bucket(bucketValidatorKeyLookups).Get(lk)
		if operatorBytes == nil {
			return errors.Errorf("no validator found for epoch %d, keyTag %d, publicKey %x: %w", epoch, keyTag, publicKey, entity.ErrEntityNotFound)
		}

		vk := epochOperatorKey(uint64(epoch), operatorBytes)
		v := tx.Bucket(bucketValidators).Get(vk)
		if v == nil {
			operator := common.BytesToAddress(operatorBytes)
			return errors.Errorf("found validator key lookup but no validator data for operator %s: %w", operator.Hex(), entity.ErrEntityNotFound)
		}

		var err error
		val, activeIndex, err = codec.BytesToValidator(v)
		return err
	})
	return val, activeIndex, err
}

func (r *Repository) GetLatestAggregatedValsetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error) {
	var epoch symbiotic.Epoch

	err := r.doView(ctx, "GetLatestAggregatedValsetHeader", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketMeta).Get(metaLatestAggregatedValsetEpoch)
		if v == nil {
			return errors.Errorf("no latest aggregated validator set found: %w", entity.ErrEntityNotFound)
		}
		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(v))
		return nil
	})
	if err != nil {
		return symbiotic.ValidatorSetHeader{}, err
	}

	return r.GetValidatorSetHeaderByEpoch(ctx, epoch)
}

func (r *Repository) saveValidatorSetMetadata(ctx context.Context, data symbiotic.ValidatorSetMetadata) error {
	metaBytes, err := codec.ValidatorSetMetadataToBytes(data)
	if err != nil {
		return errors.Errorf("failed to marshal validator set metadata: %w", err)
	}

	return r.doUpdate(ctx, "saveValidatorSetMetadata", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(data.Epoch))
		b := tx.Bucket(bucketValidatorSetMeta)
		if b.Get(ek) != nil {
			return errors.Errorf("valset metadata already exists: %w", entity.ErrEntityAlreadyExist)
		}
		return b.Put(ek, metaBytes)
	})
}

func (r *Repository) GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error) {
	var metadata symbiotic.ValidatorSetMetadata

	err := r.doView(ctx, "GetValidatorSetMetadata", func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketValidatorSetMeta).Get(epochBytes(uint64(epoch)))
		if v == nil {
			return errors.Errorf("no validatorset metadata found for epoch %v: %w", epoch, entity.ErrEntityNotFound)
		}

		var err error
		metadata, err = codec.BytesToValidatorSetMetadata(v)
		return err
	})
	return metadata, err
}
