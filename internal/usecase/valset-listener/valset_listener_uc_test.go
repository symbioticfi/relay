package valset_listener

import (
	"context"
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
		GetNetworkData(ctx, settlement).
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
		GetNetworkData(ctx, settlement).
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
	assert.Greater(t, len(result), 0)
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
	assert.Greater(t, len(result), 0)
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
