package badger

import (
	"context"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func keyNetworkConfig(epoch symbiotic.Epoch) []byte {
	return []byte(fmt.Sprintf("network_config:%d", epoch))
}

func (r *Repository) SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	configBytes, err := networkConfigToBytes(config)
	if err != nil {
		return errors.Errorf("failed to marshal network config: %w", err)
	}

	return r.doUpdateInTx(ctx, "SaveConfig", func(ctx context.Context) error {
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
	return marshalAndCompress(&v1.NetworkConfig{
		VotingPowerProviders: lo.Map(config.VotingPowerProviders, func(addr symbiotic.CrossChainAddress, _ int) *v1.CrossChainAddress {
			return &v1.CrossChainAddress{
				ChainId: addr.ChainId,
				Address: addr.Address.Bytes(),
			}
		}),
		KeysProvider: &v1.CrossChainAddress{
			Address: config.KeysProvider.Address.Bytes(),
			ChainId: config.KeysProvider.ChainId,
		},
		Settlements: lo.Map(config.Settlements, func(addr symbiotic.CrossChainAddress, _ int) *v1.CrossChainAddress {
			return &v1.CrossChainAddress{
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
		QuorumThresholds: lo.Map(config.QuorumThresholds, func(qt symbiotic.QuorumThreshold, _ int) *v1.QuorumThreshold {
			return &v1.QuorumThreshold{
				KeyTag:          uint32(qt.KeyTag),
				QuorumThreshold: qt.QuorumThreshold.String(),
			}
		}),
		NumCommitters:         config.NumCommitters,
		NumAggregators:        config.NumAggregators,
		CommitterSlotDuration: config.CommitterSlotDuration,
	})
}

func bytesToNetworkConfig(data []byte) (symbiotic.NetworkConfig, error) {
	pb := &v1.NetworkConfig{}
	if err := unmarshalAndDecompress(data, pb); err != nil {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to unmarshal network config: %w", err)
	}

	maxVotingPower, ok := new(big.Int).SetString(pb.GetMaxVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max voting power: %s", pb.GetMaxVotingPower())
	}

	minInclusionVotingPower, ok := new(big.Int).SetString(pb.GetMinInclusionVotingPower(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse min inclusion voting power: %s", pb.GetMinInclusionVotingPower())
	}

	maxValidatorsCount, ok := new(big.Int).SetString(pb.GetMaxValidatorsCount(), 10)
	if !ok {
		return symbiotic.NetworkConfig{}, errors.Errorf("failed to parse max validators count: %s", pb.GetMaxValidatorsCount())
	}

	quorumThresholds := make([]symbiotic.QuorumThreshold, 0, len(pb.GetQuorumThresholds()))

	for _, qt := range pb.GetQuorumThresholds() {
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
		VotingPowerProviders: lo.Map(pb.GetVotingPowerProviders(), func(addr *v1.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: pb.GetKeysProvider().GetChainId(),
			Address: common.BytesToAddress(pb.GetKeysProvider().GetAddress()),
		},
		Settlements: lo.Map(pb.GetSettlements(), func(addr *v1.CrossChainAddress, _ int) symbiotic.CrossChainAddress {
			return symbiotic.CrossChainAddress{
				ChainId: addr.GetChainId(),
				Address: common.BytesToAddress(addr.GetAddress()),
			}
		}),
		VerificationType:        symbiotic.VerificationType(pb.GetVerificationType()),
		MaxVotingPower:          symbiotic.ToVotingPower(maxVotingPower),
		MinInclusionVotingPower: symbiotic.ToVotingPower(minInclusionVotingPower),
		MaxValidatorsCount:      symbiotic.ToVotingPower(maxValidatorsCount),
		RequiredKeyTags:         lo.Map(pb.GetRequiredKeyTags(), func(tag uint32, _ int) symbiotic.KeyTag { return symbiotic.KeyTag(tag) }),
		RequiredHeaderKeyTag:    symbiotic.KeyTag(pb.GetRequiredHeaderKeyTag()),
		QuorumThresholds:        quorumThresholds,
		NumAggregators:          pb.GetNumAggregators(),
		NumCommitters:           pb.GetNumCommitters(),
		CommitterSlotDuration:   pb.GetCommitterSlotDuration(),
	}, nil
}
