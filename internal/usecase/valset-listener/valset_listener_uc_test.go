package valset_listener

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
)

type previousConfigRepo struct {
	repo

	config symbiotic.NetworkConfig
	saved  []entity.NextValsetData
}

func (r *previousConfigRepo) GetValidatorSetByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	return listenerTestValset(epoch), nil
}
func (r *previousConfigRepo) GetConfigByEpoch(context.Context, symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	return r.config, nil
}
func (r *previousConfigRepo) SaveNextValsetData(_ context.Context, data entity.NextValsetData) error {
	r.saved = append(r.saved, data)
	return nil
}
func (*previousConfigRepo) GetAggregationProof(context.Context, common.Hash) (symbiotic.AggregationProof, error) {
	return symbiotic.AggregationProof{}, entity.ErrEntityNotFound
}

type listenerChain struct {
	evmClient

	config symbiotic.NetworkConfig
}

func (listenerChain) GetEpochStart(context.Context, symbiotic.Epoch) (symbiotic.Timestamp, error) {
	return 0, nil
}
func (c listenerChain) GetConfig(context.Context, symbiotic.Timestamp, symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	return c.config, nil
}

type listenerDeriver struct{}

func (listenerDeriver) GetValidatorSet(_ context.Context, epoch symbiotic.Epoch, _ symbiotic.NetworkConfig) (symbiotic.ValidatorSet, error) {
	return listenerTestValset(epoch), nil
}
func (listenerDeriver) GetNetworkData(context.Context, symbiotic.CrossChainAddress) (symbiotic.NetworkData, error) {
	return symbiotic.NetworkData{Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"}}, nil
}
func listenerTestValset(epoch symbiotic.Epoch) symbiotic.ValidatorSet {
	return symbiotic.ValidatorSet{Version: 1, Epoch: epoch, QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(0))}
}

type listenerAggregator struct{ aggregator.Aggregator }

func (listenerAggregator) GenerateExtraData(context.Context, symbiotic.ValidatorSet, []symbiotic.KeyTag) ([]symbiotic.ExtraData, error) {
	return nil, nil
}

type listenerKeys struct{ keyProvider }

func (listenerKeys) GetOnchainKeyFromCache(symbiotic.KeyTag) (symbiotic.CompactPublicKey, error) {
	return nil, nil
}

func TestMissingEpochsPreservesPreviousConfig(t *testing.T) {
	previous := symbiotic.NetworkConfig{EpochDuration: 50, Settlements: []symbiotic.CrossChainAddress{{ChainId: 7}}}
	next := symbiotic.NetworkConfig{EpochDuration: 60, Settlements: []symbiotic.CrossChainAddress{{ChainId: 8}}}
	r := &previousConfigRepo{config: previous}
	s := &Service{cfg: Config{Repo: r, EvmClient: listenerChain{config: next}, Deriver: listenerDeriver{}, KeyProvider: listenerKeys{}, Aggregator: listenerAggregator{}, PollingInterval: time.Second, ValidatorSet: signals.New[symbiotic.ValidatorSet](signals.DefaultConfig(), "test")}}
	_, err := s.tryLoadMissingEpochs(t.Context(), 2, 2)
	require.NoError(t, err)
	require.Len(t, r.saved, 1)
	require.Equal(t, previous, r.saved[0].PrevNetworkConfig)
}
