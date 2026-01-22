package evm

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetChains_WithMultipleChains_ReturnsAllChainIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn1 := mocks.NewMockconn(ctrl)
	mockConn2 := mocks.NewMockconn(ctrl)
	mockConn3 := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			1:  {conn: mockConn1},
			2:  {conn: mockConn2},
			56: {conn: mockConn3},
		},
		metrics: mockMetrics,
	}

	chains := client.GetChains()

	require.Len(t, chains, 3)
	assert.Contains(t, chains, uint64(1))
	assert.Contains(t, chains, uint64(2))
	assert.Contains(t, chains, uint64(56))
}

func TestGetChains_WithNoChains_ReturnsEmptySlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chains := client.GetChains()

	require.NotNil(t, chains)
	assert.Empty(t, chains)
}

func TestGetChains_WithSingleChain_ReturnsSingleChainID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			1: {conn: mockConn},
		},
		metrics: mockMetrics,
	}

	chains := client.GetChains()

	require.Len(t, chains, 1)
	assert.Equal(t, uint64(1), chains[0])
}

func TestObserveMetrics_WithMetricsEnabled_CallsObserveEVMMethodCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("TestMethod", uint64(1), "success", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	startTime := time.Now()
	client.observeMetrics("TestMethod", 1, nil, startTime)
}

func TestObserveMetrics_WithError_CallsWithErrorStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("TestMethod", uint64(1), "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	startTime := time.Now()
	testErr := errors.New("test error")
	client.observeMetrics("TestMethod", 1, testErr, startTime)
}

func TestObserveMetrics_WithNilMetrics_DoesNotPanic(t *testing.T) {
	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: nil,
	}

	startTime := time.Now()

	require.NotPanics(t, func() {
		client.observeMetrics("TestMethod", 1, nil, startTime)
	})
}

func TestIsValsetHeaderCommittedAt_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("IsValSetHeaderCommittedAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.IsValsetHeaderCommittedAt(context.Background(), addr, 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.False(t, result)
}

func TestGetHeaderHash_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetValSetHeaderHash", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetHeaderHash(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, common.Hash{}, result)
}

func TestGetHeaderHashAt_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetValSetHeaderHashAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetHeaderHashAt(context.Background(), addr, 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, common.Hash{}, result)
}

func TestGetLastCommittedHeaderEpoch_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetLastCommittedHeaderEpoch", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetLastCommittedHeaderEpoch(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, symbiotic.Epoch(0), result)
}

func TestGetCaptureTimestampFromValsetHeaderAt_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetCaptureTimestampFromValSetHeaderAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetCaptureTimestampFromValsetHeaderAt(context.Background(), addr, 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, uint64(0), result)
}

func TestGetValSetHeaderAt_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetValSetHeaderAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetValSetHeaderAt(context.Background(), addr, 10)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, symbiotic.ValidatorSetHeader{}, result)
}

func TestGetValSetHeader_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetValSetHeader", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetValSetHeader(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, symbiotic.ValidatorSetHeader{}, result)
}

func TestGetEip712Domain_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("Eip712Domain", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetEip712Domain(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Equal(t, symbiotic.Eip712Domain{}, result)
}

func TestGetVotingPowerProviderEip712Domain_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VotingPowerProviderEip712Domain", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetVotingPowerProviderEip712Domain(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get voting power provider contract")
	assert.Equal(t, symbiotic.Eip712Domain{}, result)
}

func TestGetOperatorNonce_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetOperatorNonce", gomock.Any(), "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	operatorAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	result, err := client.GetOperatorNonce(context.Background(), addr, operatorAddr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get voting power provider contract")
	assert.Nil(t, result)
}

func TestGetVotingPowers_MulticallCheckFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetVotingPowersAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetVotingPowers(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "multicall check failed")
	assert.Nil(t, result)
}

func TestGetOperators_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetOperators", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetOperators(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create voting power provider contract")
	assert.Nil(t, result)
}

func TestGetKeysOperators_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetKeysOperators", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetKeysOperators(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create voting power provider contract")
	assert.Nil(t, result)
}

func TestGetKeys_MulticallCheckFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetKeysAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetKeys(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "multicall check failed")
	assert.Nil(t, result)
}

func TestIsValsetHeaderCommittedAtEpochs_MulticallCheckFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("IsValSetHeaderCommittedAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	epochs := []symbiotic.Epoch{10, 20, 30}
	result, err := client.IsValsetHeaderCommittedAtEpochs(context.Background(), addr, epochs)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "multicall check failed")
	assert.Nil(t, result)
}

func TestGetSubnetwork_Success_ReturnsSubnetworkHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedHash := [32]byte{1, 2, 3, 4, 5}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SUBNETWORK", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		SUBNETWORK(gomock.Any()).
		Return(expectedHash, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	hash, err := client.GetSubnetwork(context.Background())

	require.NoError(t, err)
	assert.Equal(t, common.Hash(expectedHash), hash)
}

func TestGetSubnetwork_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SUBNETWORK", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		SUBNETWORK(gomock.Any()).
		Return([32]byte{}, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	hash, err := client.GetSubnetwork(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getSubnetwork")
	assert.Equal(t, common.Hash{}, hash)
}

func TestGetNetworkAddress_Success_ReturnsNetworkAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("NETWORK", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		NETWORK(gomock.Any()).
		Return(expectedAddress, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	address, err := client.GetNetworkAddress(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expectedAddress, address)
}

func TestGetNetworkAddress_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("NETWORK", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		NETWORK(gomock.Any()).
		Return(common.Address{}, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	address, err := client.GetNetworkAddress(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getSubnetwork")
	assert.Equal(t, common.Address{}, address)
}

func TestFormatEVMContractError_WithNonJSONError_ReturnsOriginalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	originalErr := errors.New("simple error")
	result := client.formatEVMContractError(nil, originalErr)

	assert.Equal(t, originalErr, result)
}

func TestGetVotingPowers_WithNoMulticall_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetVotingPowersAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetVotingPowers(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetKeys_WithNoMulticall_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetKeysAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	result, err := client.GetKeys(context.Background(), addr, 1000)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetSettlementContract_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	result, err := client.getSettlementContract(addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}

func TestGetSettlementContract_WithConnection_ReturnsContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			chainID: {
				conn: mockConn,
			},
		},
		metrics: mockMetrics,
	}

	result, err := client.getSettlementContract(addr)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetVotingPowerProviderContract_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	result, err := client.getVotingPowerProviderContract(addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}

func TestGetVotingPowerProviderContract_WithConnection_ReturnsContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			chainID: {conn: mockConn},
		},
		metrics: mockMetrics,
	}

	result, err := client.getVotingPowerProviderContract(addr)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetVotingPowerProviderContractTransactor_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	result, err := client.getVotingPowerProviderContract(addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}

func TestGetVotingPowerProviderContractTransactor_WithConnection_ReturnsContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			chainID: {
				conn:                          mockConn,
				hasMaxPriorityFeePerGasMethod: false,
			},
		},
		metrics: mockMetrics,
	}

	result, err := client.getVotingPowerProviderContract(addr)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetKeyRegistryContract_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	result, err := client.getKeyRegistryContract(addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}

func TestGetKeyRegistryContract_WithConnection_ReturnsContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			chainID: {conn: mockConn},
		},
		metrics: mockMetrics,
	}

	result, err := client.getKeyRegistryContract(addr)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetOperatorRegistryContract_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	result, err := client.getOperatorRegistryContract(addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}

func TestGetOperatorRegistryContract_WithConnection_ReturnsContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns: map[uint64]clientWithInfo{
			chainID: {conn: mockConn},
		},
		metrics: mockMetrics,
	}

	result, err := client.getOperatorRegistryContract(addr)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGetCurrentEpoch_Success_ReturnsEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedEpoch := big.NewInt(42)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetCurrentEpoch", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		GetCurrentEpoch(gomock.Any()).
		Return(expectedEpoch, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	epoch, err := client.GetCurrentEpoch(context.Background())

	require.NoError(t, err)
	assert.Equal(t, symbiotic.Epoch(42), epoch)
}

func TestGetCurrentEpoch_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetCurrentEpoch", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetCurrentEpoch(gomock.Any()).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	epoch, err := client.GetCurrentEpoch(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getCurrentEpoch")
	assert.Equal(t, symbiotic.Epoch(0), epoch)
}

func TestGetCurrentEpochDuration_Success_ReturnsDuration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedDuration := big.NewInt(3600)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetCurrentEpochDuration", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		GetCurrentEpochDuration(gomock.Any()).
		Return(expectedDuration, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	duration, err := client.GetCurrentEpochDuration(context.Background())

	require.NoError(t, err)
	assert.Equal(t, uint64(3600), duration)
}

func TestGetCurrentEpochDuration_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetCurrentEpochDuration", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetCurrentEpochDuration(gomock.Any()).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	duration, err := client.GetCurrentEpochDuration(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getCurrentEpochDuration")
	assert.Equal(t, uint64(0), duration)
}

func TestGetEpochDuration_Success_ReturnsDuration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedDuration := big.NewInt(7200)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochDuration", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		GetEpochDuration(gomock.Any(), big.NewInt(10)).
		Return(expectedDuration, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	duration, err := client.GetEpochDuration(context.Background(), symbiotic.Epoch(10))

	require.NoError(t, err)
	assert.Equal(t, uint64(7200), duration)
}

func TestGetEpochDuration_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochDuration", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetEpochDuration(gomock.Any(), big.NewInt(10)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	duration, err := client.GetEpochDuration(context.Background(), symbiotic.Epoch(10))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getEpochDuration")
	assert.Equal(t, uint64(0), duration)
}

func TestGetEpochStart_Success_ReturnsTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedTimestamp := big.NewInt(1234567890)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochStart", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		GetEpochStart(gomock.Any(), big.NewInt(5)).
		Return(expectedTimestamp, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	timestamp, err := client.GetEpochStart(context.Background(), symbiotic.Epoch(5))

	require.NoError(t, err)
	assert.Equal(t, symbiotic.Timestamp(1234567890), timestamp)
}

func TestGetEpochStart_DriverError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochStart", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetEpochStart(gomock.Any(), big.NewInt(5)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	timestamp, err := client.GetEpochStart(context.Background(), symbiotic.Epoch(5))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getEpochStart")
	assert.Equal(t, symbiotic.Timestamp(0), timestamp)
}

func TestGetConfig_Success_ReturnsConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	driverConfig := gen.IValSetDriverConfig{
		NumAggregators: big.NewInt(5),
		NumCommitters:  big.NewInt(10),
		VotingPowerProviders: []gen.IValSetDriverCrossChainAddress{
			{ChainId: 1, Addr: common.HexToAddress("0x1111")},
			{ChainId: 2, Addr: common.HexToAddress("0x2222")},
		},
		KeysProvider: gen.IValSetDriverCrossChainAddress{
			ChainId: 1,
			Addr:    common.HexToAddress("0x3333"),
		},
		Settlements: []gen.IValSetDriverCrossChainAddress{
			{ChainId: 1, Addr: common.HexToAddress("0x4444")},
		},
		MaxVotingPower:          big.NewInt(1000),
		MinInclusionVotingPower: big.NewInt(100),
		MaxValidatorsCount:      big.NewInt(50),
		RequiredKeyTags:         []uint8{1, 2, 3},
		QuorumThresholds: []gen.IValSetDriverQuorumThreshold{
			{KeyTag: 1, QuorumThreshold: big.NewInt(66)},
		},
		RequiredHeaderKeyTag:  1,
		VerificationType:      0,
		CommitterSlotDuration: big.NewInt(10),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetConfigAt", uint64(1), "success", gomock.Any())

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochDuration", uint64(1), "success", gomock.Any())

	mockDriver.EXPECT().
		GetConfigAt(gomock.Any(), big.NewInt(1000)).
		Return(driverConfig, nil)

	mockDriver.EXPECT().
		GetEpochDuration(gomock.Any(), big.NewInt(42)).
		Return(big.NewInt(3600), nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	config, err := client.GetConfig(context.Background(), symbiotic.Timestamp(1000), symbiotic.Epoch(42))

	require.NoError(t, err)
	assert.Equal(t, uint64(5), config.NumAggregators)
	assert.Equal(t, uint64(10), config.NumCommitters)
	assert.Len(t, config.VotingPowerProviders, 2)
	assert.Equal(t, uint64(1), config.VotingPowerProviders[0].ChainId)
	assert.Equal(t, common.HexToAddress("0x1111"), config.VotingPowerProviders[0].Address)
	assert.Equal(t, common.HexToAddress("0x3333"), config.KeysProvider.Address)
	assert.Len(t, config.Settlements, 1)
	assert.Equal(t, uint64(3600), config.EpochDuration)
	assert.Len(t, config.RequiredKeyTags, 3)
	assert.Len(t, config.QuorumThresholds, 1)
	assert.Equal(t, symbiotic.KeyTag(1), config.RequiredHeaderKeyTag)
}

func TestGetConfig_GetConfigAtError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	expectedErr := errors.New("rpc error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetConfigAt", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetConfigAt(gomock.Any(), big.NewInt(1000)).
		Return(gen.IValSetDriverConfig{}, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	config, err := client.GetConfig(context.Background(), symbiotic.Timestamp(1000), symbiotic.Epoch(42))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call getConfigAt")
	assert.Equal(t, symbiotic.NetworkConfig{}, config)
}

func TestGetConfig_GetEpochDurationError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDriver := mocks.NewMockdriverContract(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	driverConfig := gen.IValSetDriverConfig{
		NumAggregators: big.NewInt(5),
		NumCommitters:  big.NewInt(10),
	}

	expectedErr := errors.New("epoch duration error")

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetConfigAt", uint64(1), "error", gomock.Any())

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("GetEpochDuration", uint64(1), "error", gomock.Any())

	mockDriver.EXPECT().
		GetConfigAt(gomock.Any(), big.NewInt(1000)).
		Return(driverConfig, nil)

	mockDriver.EXPECT().
		GetEpochDuration(gomock.Any(), big.NewInt(42)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: 1,
			},
		},
		driver:        mockDriver,
		driverChainID: 1,
		metrics:       mockMetrics,
	}

	config, err := client.GetConfig(context.Background(), symbiotic.Timestamp(1000), symbiotic.Epoch(42))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current epoch duration")
	assert.Equal(t, symbiotic.NetworkConfig{}, config)
}
