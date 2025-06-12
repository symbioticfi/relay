package badger

import (
	"encoding/json"
	"fmt"
	"math/big"

	"middleware-offchain/core/entity"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

func keyNetworkConfig(epoch uint64) []byte {
	return []byte(fmt.Sprintf("network_config:%d", epoch))
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
	return json.Marshal(networkConfigDTOFromEntity(config))
}

func bytesToNetworkConfig(data []byte) (entity.NetworkConfig, error) {
	var dto networkConfigDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.NetworkConfig{}, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return dto.ToEntity(), nil
}

func (dto networkConfigDTO) ToEntity() entity.NetworkConfig {
	return entity.NetworkConfig{
		VotingPowerProviders: lo.Map(dto.VotingPowerProviders, func(addr crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: common.HexToAddress(addr.Address),
				ChainId: addr.ChainId,
			}
		}),
		KeysProvider: entity.CrossChainAddress{
			Address: common.HexToAddress(dto.KeysProvider.Address),
			ChainId: dto.KeysProvider.ChainId,
		},
		Replicas: lo.Map(dto.Replicas, func(addr crossChainAddressDTO, _ int) entity.CrossChainAddress {
			return entity.CrossChainAddress{
				Address: common.HexToAddress(addr.Address),
				ChainId: addr.ChainId,
			}
		}),
		VerificationType:        entity.VerificationType(dto.VerificationType),
		MaxVotingPower:          dto.MaxVotingPower,
		MinInclusionVotingPower: dto.MinInclusionVotingPower,
		MaxValidatorsCount:      dto.MaxValidatorsCount,
		RequiredKeyTags:         lo.Map(dto.RequiredKeyTags, func(tag uint8, _ int) entity.KeyTag { return entity.KeyTag(tag) }),
	}
}

func networkConfigDTOFromEntity(config entity.NetworkConfig) networkConfigDTO {
	return networkConfigDTO{
		VotingPowerProviders: lo.Map(config.VotingPowerProviders, func(addr entity.CrossChainAddress, _ int) crossChainAddressDTO {
			return crossChainAddressDTO{
				Address: addr.Address.Hex(),
				ChainId: addr.ChainId,
			}
		}),
		KeysProvider: crossChainAddressDTO{
			Address: config.KeysProvider.Address.Hex(),
			ChainId: config.KeysProvider.ChainId,
		},
		Replicas: lo.Map(config.Replicas, func(addr entity.CrossChainAddress, _ int) crossChainAddressDTO {
			return crossChainAddressDTO{
				Address: addr.Address.Hex(),
				ChainId: addr.ChainId,
			}
		}),
		VerificationType:        uint32(config.VerificationType),
		MaxVotingPower:          config.MaxVotingPower,
		MinInclusionVotingPower: config.MinInclusionVotingPower,
		MaxValidatorsCount:      config.MaxValidatorsCount,
		RequiredKeyTags:         lo.Map(config.RequiredKeyTags, func(tag entity.KeyTag, _ int) uint8 { return uint8(tag) }),
	}
}
