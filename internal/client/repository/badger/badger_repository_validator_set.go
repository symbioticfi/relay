package badger

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	latestValidatorSetEpochKey        = "latest_validator_set_epoch"
	latestSignedValidatorSetEpochKey  = "latest_signed_validator_set_epoch"
	firstUncommittedValidatorSetEpoch = "first_uncommitted_validator_set_epoch"
)

// keyValidatorSetHeader returns key for validator set header
// Format: "validator_set_header:" + epoch.Bytes()
// Using epoch.Bytes() ensures proper lexicographic sorting
func keyValidatorSetHeader(epoch symbiotic.Epoch) []byte {
	key := []byte("validator_set_header:")
	key = append(key, epoch.Bytes()...)
	return key
}

// keyValidatorSetHeaderPrefix returns prefix for all validator set headers
func keyValidatorSetHeaderPrefix() []byte {
	return []byte("validator_set_header:")
}

func keyValidatorByOperator(epoch symbiotic.Epoch, operator common.Address) []byte {
	return []byte(fmt.Sprintf("validator:%d:%s", epoch, operator.Hex()))
}

func keyValidatorKeyLookup(epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKeyHash common.Hash) []byte {
	return []byte(fmt.Sprintf("validator_key_lookup:%d:%d:%s", epoch, keyTag, publicKeyHash.Hex()))
}

func keyValidatorSetStatus(epoch symbiotic.Epoch) []byte {
	return []byte(fmt.Sprintf("validator_set_status:%d", epoch))
}

func keyValidatorSetMetadata(epoch symbiotic.Epoch) []byte {
	return []byte(fmt.Sprintf("validator_set_metadata:%d", epoch))
}

func (r *Repository) SaveValidatorSetMetadata(ctx context.Context, data symbiotic.ValidatorSetMetadata) error {
	metadataBytes, err := validatorSetMetadataToBytes(data)
	if err != nil {
		return errors.Errorf("failed to marshal validator set metadata: %w", err)
	}

	return r.doUpdateInTxWithLock(ctx, "SaveValidatorSetMetadata", func(ctx context.Context) error {
		txn := getTxn(ctx)
		_, err := txn.Get(keyValidatorSetMetadata(data.Epoch))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get valset metadata: %w", err)
		}
		if err == nil {
			return errors.Errorf("valset metadata already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyValidatorSetMetadata(data.Epoch), metadataBytes)
		if err != nil {
			return errors.Errorf("failed to store valset metadata: %w", err)
		}
		return nil
	}, &r.valsetMutexMap, data.Epoch)
}

func (r *Repository) GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error) {
	var metadata symbiotic.ValidatorSetMetadata

	return metadata, r.doViewInTx(ctx, "GetValidatorSetMetadata", func(ctx context.Context) error {
		txn := getTxn(ctx)
		item, err := txn.Get(keyValidatorSetMetadata(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validatorset metadata found for epoch %v: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validatorset metadata: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validatorset metadata value: %w", err)
		}

		metadata, err = bytesToValidatorSetMetadata(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validatorset metadata: %w", err)
		}

		return nil
	})
}

func (r *Repository) SaveValidatorSet(ctx context.Context, valset symbiotic.ValidatorSet) error {
	if err := valset.Validators.CheckIsSortedByOperatorAddressAsc(); err != nil {
		return errors.Errorf("validators must be sorted by operator address ascending: %w", err)
	}

	headerBytes, err := validatorSetHeaderToBytes(valset)
	if err != nil {
		return errors.Errorf("failed to marshal validator set header: %w", err)
	}

	return r.doUpdateInTxWithLock(ctx, "SaveValidatorSet", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Check if this epoch already exists by checking the header
		headerKey := keyValidatorSetHeader(valset.Epoch)
		_, err := txn.Get(headerKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get validator set header: %w", err)
		}
		if err == nil {
			return errors.Errorf("validator set for epoch %d already exists: %w", valset.Epoch, entity.ErrEntityAlreadyExist)
		}

		// Save the validator set header for its epoch
		if err = txn.Set(headerKey, headerBytes); err != nil {
			return errors.Errorf("failed to store validator set header: %w", err)
		}

		statusKey := keyValidatorSetStatus(valset.Epoch)
		_, err = txn.Get(statusKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get validator set status key: %w", err)
		}

		statusBytes := []byte{uint8(valset.Status)}
		if err = txn.Set(statusKey, statusBytes); err != nil {
			return errors.Errorf("failed to store validator set status: %w", err)
		}

		// Check if this is a newer epoch than the latest one
		latestItem, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		shouldUpdateLatest := true
		if err == nil {
			latestValue, err := latestItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
			}
			latestEpoch := binary.BigEndian.Uint64(latestValue)
			shouldUpdateLatest = latestEpoch < uint64(valset.Epoch)
		}

		// Update latest validator set epoch only if this is a newer epoch
		if shouldUpdateLatest {
			epochBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(epochBytes, uint64(valset.Epoch))
			err = txn.Set([]byte(latestValidatorSetEpochKey), epochBytes)
			if err != nil {
				return errors.Errorf("failed to store latest validator set epoch: %w", err)
			}
		}

		// Save individual validators and their key indexes
		activeIndex := uint32(0)
		for _, validator := range valset.Validators {
			currentActiveIndex := uint32(0)
			if validator.IsActive {
				currentActiveIndex = activeIndex
				activeIndex++
			}

			validatorBytes, err := validatorToBytes(validator, currentActiveIndex)
			if err != nil {
				return errors.Errorf("failed to marshal validator: %w", err)
			}

			// Save the validator data once
			validatorKey := keyValidatorByOperator(valset.Epoch, validator.Operator)
			err = txn.Set(validatorKey, validatorBytes)
			if err != nil {
				return errors.Errorf("failed to store validator: %w", err)
			}

			// Create an index for each key that points to the validator's operator address
			for _, key := range validator.Keys {
				publicKeyHash := crypto.Keccak256Hash(key.Payload)
				keyLookup := keyValidatorKeyLookup(valset.Epoch, key.Tag, publicKeyHash)
				err = txn.Set(keyLookup, validator.Operator.Bytes())
				if err != nil {
					return errors.Errorf("failed to store validator key lookup: %w", err)
				}
			}
		}

		// Store the total active validator count for this epoch
		activeCountBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(activeCountBytes, activeIndex)
		err = txn.Set(keyActiveValidatorCount(valset.Epoch), activeCountBytes)
		if err != nil {
			return errors.Errorf("failed to store active validator count: %w", err)
		}

		return nil
	}, &r.valsetMutexMap, valset.Epoch)
}

func (r *Repository) SaveLatestSignedValidatorSetEpoch(ctx context.Context, valset symbiotic.ValidatorSet) error {
	return r.doUpdateInTxWithLock(ctx, "SaveLatestSignedValidatorSetEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		epochBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(epochBytes, uint64(valset.Epoch))
		if err := txn.Set([]byte(latestSignedValidatorSetEpochKey), epochBytes); err != nil {
			return errors.Errorf("failed to store latest validator set epoch: %w", err)
		}

		return nil
	}, &r.valsetMutexMap, valset.Epoch)
}

func (r *Repository) SaveFirstUncommittedValidatorSetEpoch(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdateInTxWithLock(ctx, "SaveFirstUncommittedValidatorSetEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		epochBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(epochBytes, uint64(epoch))
		if err := txn.Set([]byte(firstUncommittedValidatorSetEpoch), epochBytes); err != nil {
			return errors.Errorf("failed to store first uncommitted validator set epoch: %w", err)
		}

		return nil
	}, &r.valsetMutexMap, epoch)
}

func (r *Repository) UpdateValidatorSetStatus(ctx context.Context, valset symbiotic.ValidatorSet) error {
	return r.doUpdateInTxWithLock(ctx, "UpdateValidatorSetStatus", func(ctx context.Context) error {
		txn := getTxn(ctx)
		statusKey := keyValidatorSetStatus(valset.Epoch)
		_, err := txn.Get(statusKey)
		if err != nil {
			return errors.Errorf("failed to get validator set status key: %w", err)
		}

		statusBytes := []byte{uint8(valset.Status)}
		if err = txn.Set(statusKey, statusBytes); err != nil {
			return errors.Errorf("failed to store validator set status: %w", err)
		}

		return nil
	}, &r.valsetMutexMap, valset.Epoch)
}

func (r *Repository) GetValidatorSetHeaderByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetHeader, error) {
	var header symbiotic.ValidatorSetHeader

	return header, r.doViewInTx(ctx, "GetValidatorSetHeaderByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		item, err := txn.Get(keyValidatorSetHeader(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set header found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err = bytesToValidatorSetHeader(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		return nil
	})
}

func (r *Repository) getAllValidatorsByEpoch(txn *badger.Txn, epoch symbiotic.Epoch) (symbiotic.Validators, error) {
	prefix := []byte(fmt.Sprintf("validator:%d:", epoch))
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	var validators symbiotic.Validators
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		value, err := item.ValueCopy(nil)
		if err != nil {
			return nil, errors.Errorf("failed to copy validator value: %w", err)
		}

		validator, _, err := bytesToValidator(value)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal validator: %w", err)
		}

		validators = append(validators, validator)
	}

	validators.SortByOperatorAddressAsc()

	return validators, nil
}

func (r *Repository) GetValidatorSetsByEpoch(ctx context.Context, startEpoch symbiotic.Epoch) ([]symbiotic.ValidatorSet, error) {
	var validatorSets []symbiotic.ValidatorSet

	return validatorSets, r.doViewInTx(ctx, "GetValidatorSetsByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Create iterator starting from startEpoch
		startKey := keyValidatorSetHeader(startEpoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keyValidatorSetHeaderPrefix()

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(startKey); it.Valid(); it.Next() {
			item := it.Item()

			headerValue, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy validator set header value: %w", err)
			}

			header, err := bytesToValidatorSetHeader(headerValue)
			if err != nil {
				return errors.Errorf("failed to unmarshal validator set header: %w", err)
			}

			statusItem, err := txn.Get(keyValidatorSetStatus(header.Epoch))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					continue
				}
				return errors.Errorf("failed to get validator set status: %w", err)
			}

			statusValue, err := statusItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy validator set status value: %w", err)
			}

			if len(statusValue) != 1 {
				return errors.New("failed to get validator set status value: invalid length")
			}

			status := symbiotic.ValidatorSetStatus(statusValue[0])

			validators, err := r.getAllValidatorsByEpoch(txn, header.Epoch)
			if err != nil {
				return errors.Errorf("failed to get validators for epoch %d: %w", header.Epoch, err)
			}

			aggIndices, commIndices, err := extractAdditionalInfoFromHeaderData(headerValue)
			if err != nil {
				return errors.Errorf("failed to extract bitmap indices: %w", err)
			}

			validatorSets = append(validatorSets, symbiotic.ValidatorSet{
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
}

func (r *Repository) GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	var vs symbiotic.ValidatorSet

	return vs, r.doViewInTx(ctx, "GetValidatorSetByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Get the validator set header
		headerItem, err := txn.Get(keyValidatorSetHeader(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		headerValue, err := headerItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err := bytesToValidatorSetHeader(headerValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		statusItem, err := txn.Get(keyValidatorSetStatus(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set status found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set status: %w", err)
		}

		statusValue, err := statusItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set status value: %w", err)
		}

		if len(statusValue) != 1 {
			return errors.New("failed to get validator set status value: invalid length")
		}

		status := symbiotic.ValidatorSetStatus(statusValue[0])

		// Get all validators for this epoch
		validators, err := r.getAllValidatorsByEpoch(txn, epoch)
		if err != nil {
			return errors.Errorf("failed to get validators for epoch %d: %w", epoch, err)
		}

		// Extract bitmap indices from header data
		aggIndices, commIndices, err := extractAdditionalInfoFromHeaderData(headerValue)
		if err != nil {
			return errors.Errorf("failed to extract bitmap indices: %w", err)
		}

		// Build the validator set from header + validators
		vs = symbiotic.ValidatorSet{
			Version:           header.Version,
			RequiredKeyTag:    header.RequiredKeyTag,
			Epoch:             header.Epoch,
			CaptureTimestamp:  header.CaptureTimestamp,
			QuorumThreshold:   header.QuorumThreshold,
			Validators:        validators,
			Status:            status,
			AggregatorIndices: aggIndices,
			CommitterIndices:  commIndices,
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error) {
	var header symbiotic.ValidatorSetHeader

	return header, r.doViewInTx(ctx, "GetLatestValidatorSetHeader", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Get the latest epoch
		item, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
		}

		latestEpoch := symbiotic.Epoch(binary.BigEndian.Uint64(value))

		// Get the validator set header for that epoch in the same transaction
		headerItem, err := txn.Get(keyValidatorSetHeader(latestEpoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator set header found for epoch %d: %w", latestEpoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		headerValue, err := headerItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator set header value: %w", err)
		}

		header, err = bytesToValidatorSetHeader(headerValue)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator set header: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	return epoch, r.doViewInTx(ctx, "GetLatestValidatorSetEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Get the latest epoch
		item, err := txn.Get([]byte(latestValidatorSetEpochKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
		}

		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(value))
		return nil
	})
}

func keyActiveValidatorCount(epoch symbiotic.Epoch) []byte {
	return []byte(fmt.Sprintf("active_validator_count:%d", epoch))
}

func (r *Repository) GetActiveValidatorCountByEpoch(ctx context.Context, epoch symbiotic.Epoch) (uint32, error) {
	var count uint32

	return count, r.doViewInTx(ctx, "GetActiveValidatorCountByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		item, err := txn.Get(keyActiveValidatorCount(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no active validator count found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get active validator count: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy active validator count value: %w", err)
		}

		count = binary.BigEndian.Uint32(value)
		return nil
	})
}

func (r *Repository) GetLatestSignedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	return epoch, r.doViewInTx(ctx, "GetLatestSignedValidatorSetEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		item, err := txn.Get([]byte(latestSignedValidatorSetEpochKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no latest validator set found: %w", entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get latest validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy latest validator set epoch value: %w", err)
		}

		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(value))
		return nil
	})
}

func (r *Repository) GetFirstUncommittedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error) {
	var epoch symbiotic.Epoch

	return epoch, r.doViewInTx(ctx, "GetFirstUncommittedValidatorSetEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Get the latest epoch
		item, err := txn.Get([]byte(firstUncommittedValidatorSetEpoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return errors.Errorf("failed to get first uncommitted validator set epoch: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy first uncommitted validator set epoch value: %w", err)
		}

		epoch = symbiotic.Epoch(binary.BigEndian.Uint64(value))
		return nil
	})
}

func (r *Repository) GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error) {
	publicKeyHash := crypto.Keccak256Hash(publicKey)
	keyLookup := keyValidatorKeyLookup(epoch, keyTag, publicKeyHash)

	var validator symbiotic.Validator
	var activeIndex uint32
	return validator, activeIndex, r.doViewInTx(ctx, "GetValidatorByKey", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// First, find the operator address from the key lookup table
		item, err := txn.Get(keyLookup)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no validator found for epoch %d, keyTag %d, publicKey %x: %w", epoch, keyTag, publicKey, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator key lookup: %w", err)
		}

		operatorBytes, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy operator address value: %w", err)
		}
		operator := common.BytesToAddress(operatorBytes)

		// Now, retrieve the full validator data
		validatorKey := keyValidatorByOperator(epoch, operator)
		item, err = txn.Get(validatorKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				// This would indicate data inconsistency
				return errors.Errorf("found validator key lookup but no validator data for operator %s: %w", operator.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get validator data: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy validator value: %w", err)
		}

		validator, activeIndex, err = bytesToValidator(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal validator: %w", err)
		}

		return nil
	})
}

func validatorToBytes(validator symbiotic.Validator, activeIndex uint32) ([]byte, error) {
	return marshalProto(&pb.Validator{
		Operator:    validator.Operator.Bytes(),
		VotingPower: validator.VotingPower.String(),
		IsActive:    validator.IsActive,
		ActiveIndex: activeIndex,
		Keys: lo.Map(validator.Keys, func(k symbiotic.ValidatorKey, _ int) *pb.ValidatorKey {
			return &pb.ValidatorKey{
				Tag:     uint32(k.Tag),
				Payload: k.Payload,
			}
		}),
		Vaults: lo.Map(validator.Vaults, func(v symbiotic.ValidatorVault, _ int) *pb.ValidatorVault {
			return &pb.ValidatorVault{
				ChainId:     v.ChainID,
				Vault:       v.Vault.Bytes(),
				VotingPower: v.VotingPower.String(),
			}
		}),
	})
}

func bytesToValidator(data []byte) (symbiotic.Validator, uint32, error) {
	validator := &pb.Validator{}
	if err := unmarshalProto(data, validator); err != nil {
		return symbiotic.Validator{}, 0, errors.Errorf("failed to unmarshal validator: %w", err)
	}

	operator := common.BytesToAddress(validator.GetOperator())

	votingPower, ok := new(big.Int).SetString(validator.GetVotingPower(), 10)
	if !ok {
		return symbiotic.Validator{}, 0, errors.Errorf("failed to parse voting power: %s", validator.GetVotingPower())
	}

	keys := lo.Map(validator.GetKeys(), func(k *pb.ValidatorKey, _ int) symbiotic.ValidatorKey {
		return symbiotic.ValidatorKey{
			Tag:     symbiotic.KeyTag(k.GetTag()),
			Payload: k.GetPayload(),
		}
	})

	vaults := make([]symbiotic.ValidatorVault, 0, len(validator.GetVaults()))
	for _, v := range validator.GetVaults() {
		votingPowerVault, parseOk := new(big.Int).SetString(v.GetVotingPower(), 10)
		if !parseOk {
			return symbiotic.Validator{}, 0, errors.Errorf("failed to parse vault voting power for operator %s: %s", operator.Hex(), v.GetVotingPower())
		}
		vaults = append(vaults, symbiotic.ValidatorVault{
			ChainID:     v.GetChainId(),
			Vault:       common.BytesToAddress(v.GetVault()),
			VotingPower: symbiotic.ToVotingPower(votingPowerVault),
		})
	}

	return symbiotic.Validator{
		Operator:    operator,
		VotingPower: symbiotic.ToVotingPower(votingPower),
		IsActive:    validator.GetIsActive(),
		Keys:        keys,
		Vaults:      vaults,
	}, validator.GetActiveIndex(), nil
}

func validatorSetHeaderToBytes(valset symbiotic.ValidatorSet) ([]byte, error) {
	header, err := valset.GetHeader()
	if err != nil {
		return nil, errors.Errorf("failed to get validator set header: %w", err)
	}

	var aggIndices, commIndices []byte
	if len(valset.AggregatorIndices) > 0 {
		aggBitmap := entity.NewBitmapOf(valset.AggregatorIndices...)
		aggIndices, err = aggBitmap.ToBytes()
		if err != nil {
			return nil, errors.Errorf("failed to serialize aggregator indices: %w", err)
		}
	}

	if len(valset.CommitterIndices) > 0 {
		commBitmap := entity.NewBitmapOf(valset.CommitterIndices...)
		commIndices, err = commBitmap.ToBytes()
		if err != nil {
			return nil, errors.Errorf("failed to serialize committer indices: %w", err)
		}
	}

	return marshalProto(&pb.ValidatorSetHeader{
		Version:            uint32(header.Version),
		RequiredKeyTag:     uint32(header.RequiredKeyTag),
		Epoch:              uint64(header.Epoch),
		CaptureTimestamp:   uint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold.String(),
		TotalVotingPower:   header.TotalVotingPower.String(),
		ValidatorsSszMroot: header.ValidatorsSszMRoot.Bytes(),
		AggregatorIndices:  aggIndices,
		CommitterIndices:   commIndices,
	})
}

func bytesToValidatorSetHeader(data []byte) (symbiotic.ValidatorSetHeader, error) {
	validatorSetHeader := &pb.ValidatorSetHeader{}
	if err := unmarshalProto(data, validatorSetHeader); err != nil {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to unmarshal validator set header: %w", err)
	}

	quorumThreshold, ok := new(big.Int).SetString(validatorSetHeader.GetQuorumThreshold(), 10)
	if !ok {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to parse quorum threshold: %s", validatorSetHeader.GetQuorumThreshold())
	}

	totalVotingPower, ok := new(big.Int).SetString(validatorSetHeader.GetTotalVotingPower(), 10)
	if !ok {
		return symbiotic.ValidatorSetHeader{}, errors.Errorf("failed to parse total voting power: %s", validatorSetHeader.GetTotalVotingPower())
	}

	return symbiotic.ValidatorSetHeader{
		Version:            uint8(validatorSetHeader.GetVersion()),
		RequiredKeyTag:     symbiotic.KeyTag(validatorSetHeader.GetRequiredKeyTag()),
		Epoch:              symbiotic.Epoch(validatorSetHeader.GetEpoch()),
		CaptureTimestamp:   symbiotic.Timestamp(validatorSetHeader.GetCaptureTimestamp()),
		QuorumThreshold:    symbiotic.ToVotingPower(quorumThreshold),
		TotalVotingPower:   symbiotic.ToVotingPower(totalVotingPower),
		ValidatorsSszMRoot: common.BytesToHash(validatorSetHeader.GetValidatorsSszMroot()),
	}, nil
}

func extractAdditionalInfoFromHeaderData(data []byte) (aggIndices []uint32, commIndices []uint32, err error) {
	validatorSetHeader := &pb.ValidatorSetHeader{}
	if err := unmarshalProto(data, validatorSetHeader); err != nil {
		return nil, nil, errors.Errorf("failed to unmarshal validator set header: %w", err)
	}

	if len(validatorSetHeader.GetAggregatorIndices()) > 0 {
		aggBitmap, err := entity.BitmapFromBytes(validatorSetHeader.GetAggregatorIndices())
		if err != nil {
			return nil, nil, errors.Errorf("failed to deserialize aggregator indices: %w", err)
		}
		aggIndices = aggBitmap.ToArray()
	} else {
		aggIndices = []uint32{}
	}

	if len(validatorSetHeader.GetCommitterIndices()) > 0 {
		commBitmap, err := entity.BitmapFromBytes(validatorSetHeader.GetCommitterIndices())
		if err != nil {
			return nil, nil, errors.Errorf("failed to deserialize committer indices: %w", err)
		}
		commIndices = commBitmap.ToArray()
	} else {
		commIndices = []uint32{}
	}

	return aggIndices, commIndices, nil
}

func validatorSetMetadataToBytes(data symbiotic.ValidatorSetMetadata) ([]byte, error) {
	return marshalProto(&pb.ValidatorSetMetadata{
		RequestId: data.RequestID.Bytes(),
		Epoch:     uint64(data.Epoch),
		ExtraData: lo.Map(data.ExtraData, func(ed symbiotic.ExtraData, _ int) *pb.ExtraData {
			return &pb.ExtraData{
				Key:   ed.Key.Bytes(),
				Value: ed.Value.Bytes(),
			}
		}),
		CommitmentData: data.CommitmentData,
	})
}

func bytesToValidatorSetMetadata(data []byte) (symbiotic.ValidatorSetMetadata, error) {
	validatorSetMetadata := &pb.ValidatorSetMetadata{}
	if err := unmarshalProto(data, validatorSetMetadata); err != nil {
		return symbiotic.ValidatorSetMetadata{}, errors.Errorf("failed to unmarshal validator set metadata: %w", err)
	}

	return symbiotic.ValidatorSetMetadata{
		RequestID: common.BytesToHash(validatorSetMetadata.GetRequestId()),
		ExtraData: lo.Map(validatorSetMetadata.GetExtraData(), func(ed *pb.ExtraData, _ int) symbiotic.ExtraData {
			return symbiotic.ExtraData{
				Key:   common.BytesToHash(ed.GetKey()),
				Value: common.BytesToHash(ed.GetValue()),
			}
		}),
		Epoch:          symbiotic.Epoch(validatorSetMetadata.GetEpoch()),
		CommitmentData: validatorSetMetadata.GetCommitmentData(),
	}, nil
}
