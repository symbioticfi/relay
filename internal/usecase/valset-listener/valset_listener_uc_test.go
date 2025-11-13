package valset_listener

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/internal/usecase/valset-listener/mocks"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
)

type mockAggregator struct{}

func (m *mockAggregator) Aggregate(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, messageHash []byte, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error) {
	return symbiotic.AggregationProof{}, nil
}

func (m *mockAggregator) Verify(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, proof symbiotic.AggregationProof) (bool, error) {
	return true, nil
}

func (m *mockAggregator) GenerateExtraData(valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error) {
	return []symbiotic.ExtraData{}, nil
}

var _ aggregator.Aggregator = (*mockAggregator)(nil)

type failingMockAggregator struct{}

func (m *failingMockAggregator) Aggregate(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, messageHash []byte, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error) {
	return symbiotic.AggregationProof{}, errors.New("aggregate failed")
}

func (m *failingMockAggregator) Verify(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, proof symbiotic.AggregationProof) (bool, error) {
	return false, errors.New("verify failed")
}

func (m *failingMockAggregator) GenerateExtraData(valset symbiotic.ValidatorSet, keyTags []symbiotic.KeyTag) ([]symbiotic.ExtraData, error) {
	return nil, errors.New("generate extra data failed")
}

var _ aggregator.Aggregator = (*failingMockAggregator)(nil)

func TestNew_WithValidConfig_ReturnsService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmCli := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmCli,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	}

	service, err := New(cfg)

	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.cfg)
}

func TestNew_WithInvalidConfig_ReturnsError(t *testing.T) {
	cfg := Config{}

	service, err := New(cfg)

	require.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithMissingEvmClient_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       nil,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithMissingRepo_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            nil,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithMissingDeriver_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         nil,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithZeroPollingInterval_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: 0,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithMissingSigner_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          nil,
		ValidatorSet:    mockValidatorSetSignal,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithMissingValidatorSet_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    nil,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestConfigValidate_WithOptionalFieldsNil_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     nil,
		Aggregator:      nil,
		Metrics:         nil,
		ForceCommitter:  false,
	}

	err := cfg.Validate()

	require.NoError(t, err)
}

func TestConfigValidate_WithAllFieldsSet_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockAgg := &mockAggregator{}
	mockMetrics := mocks.NewMockmetrics(ctrl)

	cfg := Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
		Metrics:         mockMetrics,
		ForceCommitter:  true,
	}

	err := cfg.Validate()

	require.NoError(t, err)
}

func TestGetNetworkData_WithValidSettlement_ReturnsNetworkData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	expectedNetworkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(expectedNetworkData, nil)

	result, err := service.getNetworkData(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, expectedNetworkData, result)
}

func TestGetNetworkData_WithMultipleSettlementsFirstFails_ReturnsSecondNetworkData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement1 := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	settlement2 := symbiotic.CrossChainAddress{ChainId: 2, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement1, settlement2},
	}

	expectedNetworkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{2, 3, 4},
		Eip712Data: symbiotic.Eip712Domain{Name: "test2", Version: "2"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(ctx, settlement1).
		Return(symbiotic.NetworkData{}, errors.New("rpc error"))

	mockDeriver.EXPECT().
		GetNetworkData(ctx, settlement2).
		Return(expectedNetworkData, nil)

	result, err := service.getNetworkData(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, expectedNetworkData, result)
}

func TestGetNetworkData_WithAllSettlementsFail_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(symbiotic.NetworkData{}, errors.New("network error"))

	result, err := service.getNetworkData(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get network data for any settlement")
	assert.Equal(t, symbiotic.NetworkData{}, result)
}

func TestDetectLastCommittedEpoch_WithSingleSettlement_ReturnsEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(5), nil)

	result := service.detectLastCommittedEpoch(ctx, config)

	assert.Equal(t, symbiotic.Epoch(5), result)
}

func TestDetectLastCommittedEpoch_WithMultipleSettlements_ReturnsMinimumEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement1 := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	settlement2 := symbiotic.CrossChainAddress{ChainId: 2, Address: common.HexToAddress("0x456")}
	settlement3 := symbiotic.CrossChainAddress{ChainId: 3, Address: common.HexToAddress("0x789")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement1, settlement2, settlement3},
	}

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement1).
		Return(symbiotic.Epoch(10), nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement2).
		Return(symbiotic.Epoch(5), nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement3).
		Return(symbiotic.Epoch(8), nil)

	result := service.detectLastCommittedEpoch(ctx, config)

	assert.Equal(t, symbiotic.Epoch(5), result)
}

func TestDetectLastCommittedEpoch_WithOneSettlementError_ReturnsMinFromValidOnes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement1 := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	settlement2 := symbiotic.CrossChainAddress{ChainId: 2, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement1, settlement2},
	}

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement1).
		Return(symbiotic.Epoch(0), errors.New("rpc error"))

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement2).
		Return(symbiotic.Epoch(7), nil)

	result := service.detectLastCommittedEpoch(ctx, config)

	assert.Equal(t, symbiotic.Epoch(7), result)
}

func TestDetectLastCommittedEpoch_WithAllSettlementErrors_ReturnsZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(0), errors.New("network down"))

	result := service.detectLastCommittedEpoch(ctx, config)

	assert.Equal(t, symbiotic.Epoch(0), result)
}

func TestCommitValsetToAllSettlements_WhenAlreadyCommitted_SkipsCommit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 5}
	extraData := []symbiotic.ExtraData{}
	proof := []byte{1, 2, 3}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch).
		Return(true, nil)

	err = service.commitValsetToAllSettlements(ctx, config, header, extraData, proof)

	require.NoError(t, err)
}

func TestCommitValsetToAllSettlements_WhenCheckFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 5}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch).
		Return(false, errors.New("rpc error"))

	err = service.commitValsetToAllSettlements(ctx, config, header, []symbiotic.ExtraData{}, []byte{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if header is committed")
}

func TestCommitValsetToAllSettlements_WhenEpochNotSequential_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 10}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch).
		Return(false, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(5), nil)

	err = service.commitValsetToAllSettlements(ctx, config, header, []symbiotic.ExtraData{}, []byte{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "commits should be consequent")
}

func TestHeaderCommitmentData_WithValidInputs_ReturnsCommitmentData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{
			Name:    "TestNetwork",
			Version: "1.0",
		},
	}

	header := symbiotic.ValidatorSetHeader{
		Epoch:            10,
		CaptureTimestamp: 1234567890,
	}

	extraData := []symbiotic.ExtraData{
		{Key: common.HexToHash("0x01"), Value: common.HexToHash("0x02")},
	}

	result, err := service.headerCommitmentData(networkData, header, extraData)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestHeaderCommitmentData_WithEmptyExtraData_ReturnsCommitmentData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{5, 6, 7},
		Eip712Data: symbiotic.Eip712Domain{
			Name:    "TestNet",
			Version: "2.0",
		},
	}

	header := symbiotic.ValidatorSetHeader{
		Epoch:            5,
		CaptureTimestamp: 9876543210,
	}

	result, err := service.headerCommitmentData(networkData, header, []symbiotic.ExtraData{})

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestTryLoadMissingEpochs_WithLatestEpochOngoing_ReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()

	header := symbiotic.ValidatorSetHeader{
		Epoch:            5,
		CaptureTimestamp: symbiotic.Timestamp(time.Now().Add(1 * time.Hour).Unix()),
	}

	config := symbiotic.NetworkConfig{
		EpochDuration: 7200,
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(ctx).
		Return(header, nil)

	mockRepo.EXPECT().
		GetConfigByEpoch(ctx, header.Epoch).
		Return(config, nil)

	err = service.tryLoadMissingEpochs(ctx)

	require.NoError(t, err)
}

func TestTryLoadMissingEpochs_WhenGetLatestHeaderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(ctx).
		Return(symbiotic.ValidatorSetHeader{}, errors.New("database error"))

	err = service.tryLoadMissingEpochs(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get latest validator set header")
}

func TestTryLoadMissingEpochs_WhenGetCurrentEpochFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(ctx).
		Return(symbiotic.ValidatorSetHeader{}, entity.ErrEntityNotFound)

	mockEvmClient.EXPECT().
		GetCurrentEpoch(ctx).
		Return(symbiotic.Epoch(0), errors.New("rpc connection failed"))

	err = service.tryLoadMissingEpochs(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current epoch")
}

func TestLoadAllMissingEpochs_WhenTryLoadFails_RetriesAndFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(gomock.Any()).
		Return(symbiotic.ValidatorSetHeader{}, errors.New("database error")).
		Times(10)

	err = service.LoadAllMissingEpochs(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load missing epochs after")
}

func TestProcess_WhenGetNetworkDataFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	valSet := symbiotic.ValidatorSet{
		Epoch: 5,
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(symbiotic.NetworkData{}, errors.New("network error"))

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get network data")
}

func TestProcess_WhenGenerateExtraDataFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &failingMockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate extra data")
}

func TestProcess_WhenSaveMetadataFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{0},
	}
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockRepo.EXPECT().
		SaveValidatorSetMetadata(gomock.Any(), gomock.Any()).
		Return(errors.New("database write error"))

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save validator set metadata")
}

func TestProcess_WhenSaveProofCommitPendingAlreadyExists_ReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{0},
	}
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockRepo.EXPECT().
		SaveValidatorSetMetadata(gomock.Any(), gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		SaveProofCommitPending(gomock.Any(), valSet.Epoch, gomock.Any()).
		Return(entity.ErrEntityAlreadyExist)

	err = service.process(ctx, valSet, config)

	require.NoError(t, err)
}

func TestCommitValsetToAllSettlements_WhenCommitSucceeds_ReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 5}
	extraData := []symbiotic.ExtraData{{Key: common.HexToHash("0x01"), Value: common.HexToHash("0x02")}}
	proof := []byte{1, 2, 3}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch).
		Return(false, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(4), nil)

	mockEvmClient.EXPECT().
		CommitValsetHeader(ctx, settlement, header, extraData, proof).
		Return(symbiotic.TxResult{TxHash: common.HexToHash("0xabc")}, nil)

	err = service.commitValsetToAllSettlements(ctx, config, header, extraData, proof)

	require.NoError(t, err)
}

func TestCommitValsetToAllSettlements_WhenCommitFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 5}
	extraData := []symbiotic.ExtraData{{Key: common.HexToHash("0x01"), Value: common.HexToHash("0x02")}}
	proof := []byte{1, 2, 3}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch).
		Return(false, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(4), nil)

	mockEvmClient.EXPECT().
		CommitValsetHeader(ctx, settlement, header, extraData, proof).
		Return(symbiotic.TxResult{}, errors.New("transaction failed"))

	err = service.commitValsetToAllSettlements(ctx, config, header, extraData, proof)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit valset header to settlement")
}

func TestCommitValsetToAllSettlements_WithMultipleSettlements_ReturnsPartialErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement1 := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	settlement2 := symbiotic.CrossChainAddress{ChainId: 2, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement1, settlement2},
	}
	header := symbiotic.ValidatorSetHeader{Epoch: 5}
	extraData := []symbiotic.ExtraData{{Key: common.HexToHash("0x01"), Value: common.HexToHash("0x02")}}
	proof := []byte{1, 2, 3}

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement1, header.Epoch).
		Return(false, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement1).
		Return(symbiotic.Epoch(4), nil)

	mockEvmClient.EXPECT().
		CommitValsetHeader(ctx, settlement1, header, extraData, proof).
		Return(symbiotic.TxResult{TxHash: common.HexToHash("0xabc")}, nil)

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(ctx, settlement2, header.Epoch).
		Return(false, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement2).
		Return(symbiotic.Epoch(4), nil)

	mockEvmClient.EXPECT().
		CommitValsetHeader(ctx, settlement2, header, extraData, proof).
		Return(symbiotic.TxResult{}, errors.New("rpc error"))

	err = service.commitValsetToAllSettlements(ctx, config, header, extraData, proof)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc error")
}

func TestProcess_WhenIsSignerTrue_CallsRequestSignature(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			Keys: []symbiotic.ValidatorKey{{
				Tag:     keyTag,
				Payload: onchainKey,
			}},
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			Keys: []symbiotic.ValidatorKey{{
				Tag:     keyTag,
				Payload: onchainKey,
			}},
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockSigner.EXPECT().
		RequestSignature(gomock.Any(), gomock.Any()).
		Return(common.HexToHash("0x123"), nil)

	mockRepo.EXPECT().
		SaveValidatorSetMetadata(gomock.Any(), gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		SaveProofCommitPending(gomock.Any(), valSet.Epoch, gomock.Any()).
		Return(nil)

	err = service.process(ctx, valSet, config)

	require.NoError(t, err)
}

func TestProcess_WhenGenesisEpoch_DoesNotGetPreviousValset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}
	valSet := symbiotic.ValidatorSet{
		Epoch:          0,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockRepo.EXPECT().
		SaveValidatorSetMetadata(gomock.Any(), gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		SaveProofCommitPending(gomock.Any(), valSet.Epoch, gomock.Any()).
		Return(nil)

	err = service.process(ctx, valSet, config)

	require.NoError(t, err)
}

func TestProcess_WhenFullSuccessPath_SavesMetadataAndPendingProof(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockRepo.EXPECT().
		SaveValidatorSetMetadata(gomock.Any(), gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		SaveProofCommitPending(gomock.Any(), valSet.Epoch, gomock.Any()).
		Return(nil)

	err = service.process(ctx, valSet, config)

	require.NoError(t, err)
}

func TestProcess_WhenGetPreviousValidatorSetFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(symbiotic.ValidatorSet{}, errors.New("database error"))

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get previous validator set")
}

func TestProcess_WhenGetOnchainKeyFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(symbiotic.CompactPublicKey{}, errors.New("key not found"))

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get onchain symb key from cache")
}

func TestProcess_WhenRequestSignatureFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x123")}
	keyTag := symbiotic.KeyTag(0)
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	onchainKey := symbiotic.CompactPublicKey{1, 2, 3}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			Keys: []symbiotic.ValidatorKey{{
				Tag:     keyTag,
				Payload: onchainKey,
			}},
		}},
	}

	prevValSet := symbiotic.ValidatorSet{
		Epoch:          4,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			Keys: []symbiotic.ValidatorKey{{
				Tag:     keyTag,
				Payload: onchainKey,
			}},
		}},
	}

	networkData := symbiotic.NetworkData{
		Subnetwork: [32]byte{1, 2, 3},
		Eip712Data: symbiotic.Eip712Domain{Name: "test", Version: "1"},
	}

	mockDeriver.EXPECT().
		GetNetworkData(gomock.Any(), settlement).
		Return(networkData, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), symbiotic.Epoch(4)).
		Return(prevValSet, nil)

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	mockSigner.EXPECT().
		RequestSignature(gomock.Any(), gomock.Any()).
		Return(common.Hash{}, errors.New("signing failed"))

	err = service.process(ctx, valSet, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sign new validator set extra")
}

func TestShouldCommitForValset_WhenForceCommitterEnabled_ReturnsTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
		ForceCommitter:  true,
	})
	require.NoError(t, err)

	ctx := context.Background()
	valSet := symbiotic.ValidatorSet{Epoch: 5}
	nwCfg := symbiotic.NetworkConfig{}

	shouldCommit, err := service.shouldCommitForValset(ctx, valSet, nwCfg)

	require.NoError(t, err)
	assert.True(t, shouldCommit)
}

func TestShouldCommitForValset_WhenKeyNotFound_ReturnsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
	}
	nwCfg := symbiotic.NetworkConfig{}

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(symbiotic.CompactPublicKey{}, entity.ErrKeyNotFound)

	shouldCommit, err := service.shouldCommitForValset(ctx, valSet, nwCfg)

	require.NoError(t, err)
	assert.False(t, shouldCommit)
}

func TestShouldCommitForValset_WhenGetOnchainKeyFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
	}
	nwCfg := symbiotic.NetworkConfig{}

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(symbiotic.CompactPublicKey{}, errors.New("key provider error"))

	shouldCommit, err := service.shouldCommitForValset(ctx, valSet, nwCfg)

	require.Error(t, err)
	assert.False(t, shouldCommit)
	assert.Contains(t, err.Error(), "failed to get onchain key")
}

func TestShouldCommitForValset_WhenNotActiveCommitter_ReturnsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	keyTag := symbiotic.KeyTag(0)
	onchainKey := symbiotic.CompactPublicKey{9, 9, 9}
	valSet := symbiotic.ValidatorSet{
		Epoch:          5,
		RequiredKeyTag: keyTag,
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x456"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			Keys: []symbiotic.ValidatorKey{{
				Tag:     keyTag,
				Payload: symbiotic.CompactPublicKey{1, 2, 3},
			}},
		}},
	}
	nwCfg := symbiotic.NetworkConfig{CommitterSlotDuration: 100}

	mockKeyProvider.EXPECT().
		GetOnchainKeyFromCache(keyTag).
		Return(onchainKey, nil)

	shouldCommit, err := service.shouldCommitForValset(ctx, valSet, nwCfg)

	require.NoError(t, err)
	assert.False(t, shouldCommit)
}

func TestProcessPendingProofCommit_WhenSucceeds_RemovesPendingState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	requestID := common.HexToHash("0x123")
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: requestID,
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{
		Proof: []byte{1, 2, 3},
	}

	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:            epoch,
		RequiredKeyTag:   keyTag,
		CaptureTimestamp: 1000,
		Validators:       []symbiotic.Validator{},
		CommitterIndices: []uint32{},
	}

	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), requestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(valSet, nil)

	mockEvmClient.EXPECT().
		GetConfig(gomock.Any(), symbiotic.Timestamp(1000), epoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(gomock.Any(), settlement, epoch).
		Return(true, nil)

	mockRepo.EXPECT().
		RemoveProofCommitPending(gomock.Any(), epoch, requestID).
		Return(nil)

	err = service.processPendingProofCommit(ctx, proofKey)

	require.NoError(t, err)
}

func TestProcessPendingProofCommit_WhenProofNotFound_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	proofKey := symbiotic.ProofCommitKey{
		RequestID: common.HexToHash("0x123"),
		Epoch:     5,
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
}

func TestProcessPendingProofCommit_WhenGetValidatorSetFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: common.HexToHash("0x123"),
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{Proof: []byte{1, 2, 3}}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(symbiotic.ValidatorSet{}, errors.New("db error"))

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestProcessPendingProofCommit_WhenGetConfigFails_ReturnsCriticalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: common.HexToHash("0x123"),
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{Proof: []byte{1, 2, 3}}
	valSet := symbiotic.ValidatorSet{
		Epoch:            epoch,
		CaptureTimestamp: 1000,
		Validators:       []symbiotic.Validator{},
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(valSet, nil)

	mockEvmClient.EXPECT().
		GetConfig(gomock.Any(), symbiotic.Timestamp(1000), epoch).
		Return(symbiotic.NetworkConfig{}, errors.New("evm error"))

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	var critErr *criticalError
	assert.True(t, errors.As(err, &critErr))
	assert.Contains(t, err.Error(), "failed to get config")
}

func TestProcessPendingProofCommit_WhenGenerateExtraDataFails_ReturnsCriticalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &failingMockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: common.HexToHash("0x123"),
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{Proof: []byte{1, 2, 3}}
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:            epoch,
		CaptureTimestamp: 1000,
		RequiredKeyTag:   keyTag,
		Validators:       []symbiotic.Validator{},
	}

	config := symbiotic.NetworkConfig{
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(valSet, nil)

	mockEvmClient.EXPECT().
		GetConfig(gomock.Any(), symbiotic.Timestamp(1000), epoch).
		Return(config, nil)

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	var critErr *criticalError
	assert.True(t, errors.As(err, &critErr))
	assert.Contains(t, err.Error(), "failed to generate extra data")
}

func TestProcessPendingProofCommit_WhenCommitFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: common.HexToHash("0x123"),
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{Proof: []byte{1, 2, 3}}
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:            epoch,
		CaptureTimestamp: 1000,
		RequiredKeyTag:   keyTag,
		Validators:       []symbiotic.Validator{},
		CommitterIndices: []uint32{},
	}

	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(valSet, nil)

	mockEvmClient.EXPECT().
		GetConfig(gomock.Any(), symbiotic.Timestamp(1000), epoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(gomock.Any(), settlement, epoch).
		Return(false, errors.New("rpc error"))

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc error")
}

func TestProcessPendingProofCommit_WhenRemovePendingFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockIEvmClient(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockDeriver := mocks.NewMockderiver(ctrl)
	mockSigner := mocks.NewMocksigner(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)
	mockValidatorSetSignal := signals.New[symbiotic.ValidatorSet](signals.Config{}, "test-valset")
	mockAgg := &mockAggregator{}

	service, err := New(Config{
		EvmClient:       mockEvmClient,
		Repo:            mockRepo,
		Deriver:         mockDeriver,
		PollingInterval: time.Second * 10,
		Signer:          mockSigner,
		ValidatorSet:    mockValidatorSetSignal,
		KeyProvider:     mockKeyProvider,
		Aggregator:      mockAgg,
	})
	require.NoError(t, err)

	ctx := context.Background()
	requestID := common.HexToHash("0x123")
	epoch := symbiotic.Epoch(5)
	proofKey := symbiotic.ProofCommitKey{
		RequestID: requestID,
		Epoch:     epoch,
	}

	proof := symbiotic.AggregationProof{Proof: []byte{1, 2, 3}}
	keyTag := symbiotic.KeyTag(0)
	valSet := symbiotic.ValidatorSet{
		Epoch:            epoch,
		CaptureTimestamp: 1000,
		RequiredKeyTag:   keyTag,
		Validators:       []symbiotic.Validator{},
		CommitterIndices: []uint32{},
	}

	settlement := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x456")}
	config := symbiotic.NetworkConfig{
		Settlements:     []symbiotic.CrossChainAddress{settlement},
		RequiredKeyTags: []symbiotic.KeyTag{keyTag},
	}

	mockRepo.EXPECT().
		GetAggregationProof(gomock.Any(), proofKey.RequestID).
		Return(proof, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(gomock.Any(), epoch).
		Return(valSet, nil)

	mockEvmClient.EXPECT().
		GetConfig(gomock.Any(), symbiotic.Timestamp(1000), epoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		IsValsetHeaderCommittedAt(gomock.Any(), settlement, epoch).
		Return(true, nil)

	mockRepo.EXPECT().
		RemoveProofCommitPending(gomock.Any(), epoch, requestID).
		Return(errors.New("db remove error"))

	err = service.processPendingProofCommit(ctx, proofKey)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db remove error")
}
