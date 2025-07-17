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

	"github.com/symbiotic/relay/core/entity"
)

func keyNetworkConfig(epoch uint64) []byte {
	return []byte(fmt.Sprintf("network_config:%d", epoch))
}

func (r *Repository) SaveConfig(_ context.Context, config entity.NetworkConfig, epoch uint64) error {
	configBytes, err := networkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(keyNetworkConfig(epoch))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get network config: %w", err)
		}
		if err == nil {
			return errors.Errorf("network config for epoch %d already exists: %w", epoch, entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyNetworkConfig(epoch), configBytes)
		if err != nil {
			return errors.Errorf("failed to store network config: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetConfigByEpoch(_ context.Context, epoch uint64) (entity.NetworkConfig, error) {
	var config entity.NetworkConfig

	return config, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyNetworkConfig(epoch))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no network config found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get network config: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy network config value: %w", err)
		}

		config, err = bytesToNetworkConfig(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal network config: %w", err)
		}

		return nil
	})
}

type crossChainAddressDTO struct {
	Address string `json:"address"`
	ChainId uint64 `json:"chain_id"`
}

type networkConfigDTO struct {
	VotingPowerProviders    []crossChainAddressDTO `json:"voting_power_providers"`
	KeysProvider            crossChainAddressDTO   `json:"keys_provider"`
	Replicas                []crossChainAddressDTO `json:"replicas"`
	VerificationType        uint32                 `json:"verification_type"`
	MaxVotingPower          *big.Int               `json:"max_voting_power"`
	MinInclusionVotingPower *big.Int               `json:"min_inclusion_voting_power"`
	MaxValidatorsCount      *big.Int               `json:"max_validators_count"`
	RequiredKeyTags         []uint8                `json:"required_key_tags"`
}

func networkConfigToBytes(config entity.NetworkConfig) ([]byte, error) {
	networkConfigDTOFromEntity := networkConfigDTO{
		VotingPowerProviders: lo.Map(config.VotingPowerProviders, func(addr entity.CrossChainAddress, _ int) crossChainAddressDTO {
			return crossChainAddressDTO{
				ChainId: addr.ChainId,
				Address: addr.Address.Hex(),
			}
		}),
		KeysProvider: crossChainAddressDTO{
			Address: config.KeysProvider.Address.Hex(),
			ChainId: config.KeysProvider.ChainId,
		},
		Replicas: lo.Map(config.Replicas, func(addr entity.CrossChainAddress, _ int) crossChainAddressDTO {
			return crossChainAddressDTO{
				ChainId: addr.ChainId,
				Address: addr.Address.Hex(),
			}
		}),
		VerificationType:        uint32(config.VerificationType),
		MaxVotingPower:          config.MaxVotingPower.Int,
		MinInclusionVotingPower: config.MinInclusionVotingPower.Int,
		MaxValidatorsCount:      config.MaxValidatorsCount.Int,
		RequiredKeyTags:         lo.Map(config.RequiredKeyTags, func(tag entity.KeyTag, _ int) uint8 { return uint8(tag) }),
	}

	return json.Marshal(networkConfigDTOFromEntity)
}

func bytesToNetworkConfig(data []byte) (entity.NetworkConfig, error) {
	var dto networkConfigDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.NetworkConfig{}, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return entity.NetworkConfig{
		VotingPowerProviders: lo.Map(dto.VotingPowerProviders, func(addr crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: common.HexToAddress(addr.Address),
			}
		}),
		KeysProvider: entity.CrossChainAddress{
			ChainId: dto.KeysProvider.ChainId,
			Address: common.HexToAddress(dto.KeysProvider.Address),
		},
		Replicas: lo.Map(dto.Replicas, func(addr crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: common.HexToAddress(addr.Address),
			}
		}),
		VerificationType:        entity.VerificationType(dto.VerificationType),
		MaxVotingPower:          entity.ToVotingPower(dto.MaxVotingPower),
		MinInclusionVotingPower: entity.ToVotingPower(dto.MinInclusionVotingPower),
		MaxValidatorsCount:      entity.ToVotingPower(dto.MaxValidatorsCount),
		RequiredKeyTags:         lo.Map(dto.RequiredKeyTags, func(tag uint8, _ int) entity.KeyTag { return entity.KeyTag(tag) }),
	}, nil
}
