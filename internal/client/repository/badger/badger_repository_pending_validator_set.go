package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
)

func keyPendingValidatorSet(reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("pending_validator_set:%s", reqHash.Hex()))
}

func (r *Repository) SavePendingValidatorSet(_ context.Context, reqHash common.Hash, vs entity.ValidatorSet) error {
	bytes, err := validatorSetToBytes(vs)
	if err != nil {
		return errors.Errorf("failed to marshal validator set: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(keyPendingValidatorSet(reqHash))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get pending validator set: %w", err)
		}
		if err == nil {
			return errors.Errorf("pending validator set already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyPendingValidatorSet(reqHash), bytes)
		if err != nil {
			return errors.Errorf("failed to store pending validator set: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetPendingValidatorSet(_ context.Context, reqHash common.Hash) (entity.ValidatorSet, error) {
	var vs entity.ValidatorSet

	return vs, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyPendingValidatorSet(reqHash))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no pending validator set found for hash %s: %w", reqHash.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get pending validator set: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy pending validator set value: %w", err)
		}

		vs, err = bytesToValidatorSet(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal pending validator set: %w", err)
		}

		return nil
	})
}

type validatorVaultDTO struct {
	ChainID     uint64 `json:"chain_id"`
	Vault       string `json:"vault"`
	VotingPower string `json:"voting_power"`
}

type keyDTO struct {
	Tag     uint8  `json:"tag"`
	Payload []byte `json:"payload"`
}

type validatorDTO struct {
	Operator    string              `json:"operator"`
	VotingPower string              `json:"voting_power"`
	IsActive    bool                `json:"is_active"`
	Keys        []keyDTO            `json:"keys"`
	Vaults      []validatorVaultDTO `json:"vaults"`
}

type validatorSetDTO struct {
	Version            uint8          `json:"version"`
	RequiredKeyTag     uint8          `json:"required_key_tag"`
	Epoch              uint64         `json:"epoch"`
	CaptureTimestamp   uint64         `json:"capture_timestamp"`
	QuorumThreshold    string         `json:"quorum_threshold"`
	PreviousHeaderHash string         `json:"previous_header_hash"`
	Validators         []validatorDTO `json:"validators"`
	Status             int            `json:"status"`
}

func validatorSetToBytes(vs entity.ValidatorSet) ([]byte, error) {
	dto := validatorSetDTO{
		Version:            vs.Version,
		RequiredKeyTag:     uint8(vs.RequiredKeyTag),
		Epoch:              vs.Epoch,
		CaptureTimestamp:   vs.CaptureTimestamp,
		QuorumThreshold:    vs.QuorumThreshold.String(),
		PreviousHeaderHash: vs.PreviousHeaderHash.Hex(),
		Validators: lo.Map(vs.Validators, func(v entity.Validator, _ int) validatorDTO {
			return validatorDTO{
				Operator:    v.Operator.Hex(),
				VotingPower: v.VotingPower.String(),
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k entity.ValidatorKey, _ int) keyDTO {
					return keyDTO{
						Tag:     uint8(k.Tag),
						Payload: k.Payload,
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v entity.ValidatorVault, _ int) validatorVaultDTO {
					return validatorVaultDTO{
						ChainID:     v.ChainID,
						Vault:       v.Vault.Hex(),
						VotingPower: v.VotingPower.String(),
					}
				}),
			}
		}),
		Status: int(vs.Status),
	}

	return json.Marshal(dto)
}

func bytesToValidatorSet(data []byte) (entity.ValidatorSet, error) {
	var dto validatorSetDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.ValidatorSet{}, fmt.Errorf("failed to unmarshal validator set: %w", err)
	}

	quorumThreshold, ok := new(big.Int).SetString(dto.QuorumThreshold, 10)
	if !ok {
		return entity.ValidatorSet{}, fmt.Errorf("failed to parse quorum threshold: %s", dto.QuorumThreshold)
	}

	return entity.ValidatorSet{
		Version:            dto.Version,
		RequiredKeyTag:     entity.KeyTag(dto.RequiredKeyTag),
		Epoch:              dto.Epoch,
		CaptureTimestamp:   dto.CaptureTimestamp,
		QuorumThreshold:    entity.ToVotingPower(quorumThreshold),
		PreviousHeaderHash: common.HexToHash(dto.PreviousHeaderHash),
		Validators: lo.Map(dto.Validators, func(v validatorDTO, _ int) entity.Validator {
			votingPower, _ := new(big.Int).SetString(v.VotingPower, 10)
			return entity.Validator{
				Operator:    common.HexToAddress(v.Operator),
				VotingPower: entity.ToVotingPower(votingPower),
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k keyDTO, _ int) entity.ValidatorKey {
					return entity.ValidatorKey{
						Tag:     entity.KeyTag(k.Tag),
						Payload: k.Payload,
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v validatorVaultDTO, _ int) entity.ValidatorVault {
					vaultVotingPower, _ := new(big.Int).SetString(v.VotingPower, 10)
					return entity.ValidatorVault{
						ChainID:     v.ChainID,
						Vault:       common.HexToAddress(v.Vault),
						VotingPower: entity.ToVotingPower(vaultVotingPower),
					}
				}),
			}
		}),
		Status: entity.ValidatorSetStatus(dto.Status),
	}, nil
}
