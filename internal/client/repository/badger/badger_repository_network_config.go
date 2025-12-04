package badger

import (
	"context"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	pb "github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const networkConfigPrefix = "network_config:"

func keyNetworkConfig(epoch symbiotic.Epoch) []byte {
	return epochKey(networkConfigPrefix, epoch)
}

func (r *Repository) saveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	configBytes, err := networkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	return r.doUpdateInTx(ctx, "saveConfig", func(ctx context.Context) error {
		txn := getTxn(ctx)
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

func (r *Repository) GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	var config symbiotic.NetworkConfig

	return config, r.doViewInTx(ctx, "GetConfigByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
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

func networkConfigToBytes(config symbiotic.NetworkConfig) ([]byte, error) {
	return marshalProto(&pb.NetworkConfig{
		VotingPowerProviders: lo.Map(config.VotingPowerProviders, func(addr symbiotic.CrossChainAddress, _ int) *pb.CrossChainAddress {
			return &pb.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: addr.Address.Bytes(),
			}
		}),
		KeysProvider: &pb.CrossChainAddress{
			Address: config.KeysProvider.Address.Bytes(),
			ChainId: config.KeysProvider.ChainId,
		},
		Settlements: lo.Map(config.Settlements, func(addr symbiotic.CrossChainAddress, _ int) *pb.CrossChainAddress {
			return &pb.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: addr.Address.Bytes(),
			}
		}),
		VerificationType:        uint32(config.VerificationType),
		MaxVotingPower:          config.MaxVotingPower.String(),
		MinInclusionVotingPower: config.MinInclusionVotingPower.String(),
		MaxValidatorsCount:      config.MaxValidatorsCount.String(),
		RequiredKeyTags:         lo.Map(config.RequiredKeyTags, func(tag symbiotic.KeyTag, _ int) uint32 { return uint32(tag) }),
		RequiredHeaderKeyTag:    uint32(config.RequiredHeaderKeyTag),
		QuorumThresholds: lo.Map(config.QuorumThresholds, func(qt symbiotic.QuorumThreshold, _ int) *pb.QuorumThreshold {
			return &pb.QuorumThreshold{
				KeyTag:          uint32(qt.KeyTag),
				QuorumThreshold: qt.QuorumThreshold.String(),
			}
		}),
		NumCommitters:         config.NumCommitters,
		NumAggregators:        config.NumAggregators,
		CommitterSlotDuration: config.CommitterSlotDuration,
		EpochDuration:         config.EpochDuration,
	})
}

func bytesToNetworkConfig(data []byte) (symbiotic.NetworkConfig, error) {
	networkConfig := &pb.NetworkConfig{}
	if err := unmarshalProto(data, networkConfig); err != nil {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to unmarshal network config: %w", err)
	}

	maxVotingPower, ok := new(big.Int).SetString(networkConfig.GetMaxVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max voting power: %s", networkConfig.GetMaxVotingPower())
	}

	minInclusionVotingPower, ok := new(big.Int).SetString(networkConfig.GetMinInclusionVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse min inclusion voting power: %s", networkConfig.GetMinInclusionVotingPower())
	}

	maxValidatorsCount, ok := new(big.Int).SetString(networkConfig.GetMaxValidatorsCount(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max validators count: %s", networkConfig.GetMaxValidatorsCount())
	}

	quorumThresholds := make([]symbiotic.QuorumThreshold, 0, len(networkConfig.GetQuorumThresholds()))

	for _, qt := range networkConfig.GetQuorumThresholds() {
		threshold, parseOk := new(big.Int).SetString(qt.GetQuorumThreshold(), 10)
		if !parseOk {
			return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse quorum threshold: %s", qt.GetQuorumThreshold())
		}

		quorumThresholds = append(quorumThresholds, symbiotic.QuorumThreshold{
			KeyTag:          symbiotic.KeyTag(qt.GetKeyTag()),
			QuorumThreshold: symbiotic.ToQuorumThresholdPct(threshold),
		})
	}

	return symbiotic.NetworkConfig{
		VotingPowerProviders: lo.Map(networkConfig.GetVotingPowerProviders(), func(addr *pb.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: networkConfig.GetKeysProvider().GetChainId(),
			Address: common.BytesToAddress(networkConfig.GetKeysProvider().GetAddress()),
		},
		Settlements: lo.Map(networkConfig.GetSettlements(), func(addr *pb.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		VerificationType:        symbiotic.VerificationType(networkConfig.GetVerificationType()),
		MaxVotingPower:          symbiotic.ToVotingPower(maxVotingPower),
		MinInclusionVotingPower: symbiotic.ToVotingPower(minInclusionVotingPower),
		MaxValidatorsCount:      symbiotic.ToVotingPower(maxValidatorsCount),
		RequiredKeyTags:         lo.Map(networkConfig.GetRequiredKeyTags(), func(tag uint32, _ int) symbiotic.KeyTag { return symbiotic.KeyTag(tag) }),
		RequiredHeaderKeyTag:    symbiotic.KeyTag(networkConfig.GetRequiredHeaderKeyTag()),
		QuorumThresholds:        quorumThresholds,
		NumAggregators:          networkConfig.GetNumAggregators(),
		NumCommitters:           networkConfig.GetNumCommitters(),
		CommitterSlotDuration:   networkConfig.GetCommitterSlotDuration(),
		EpochDuration:           networkConfig.GetEpochDuration(),
	}, nil
}
