package evm

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestVerifyQuorumSig_NoSettlementContract_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(10),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.False(t, result)
}

func TestVerifyQuorumSig_ContextTimeout_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		cfg: Config{
			RequestTimeout: 1 * time.Millisecond,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		ctx,
		addr,
		symbiotic.Epoch(10),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.False(t, result)
}

func TestVerifyQuorumSig_InvalidChainID_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(10),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.False(t, result)
}

func TestVerifyQuorumSig_EmptyProof_HandlesCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(10),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte{},
	)

	require.Error(t, err)
	assert.False(t, result)
}

func TestVerifyQuorumSig_EmptyMessage_HandlesCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(10),
		[]byte{},
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.False(t, result)
}

func TestVerifyQuorumSig_NilThreshold_HandlesCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(10),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		nil,
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.False(t, result)
}

func TestVerifyQuorumSig_ZeroEpoch_HandlesCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("VerifyQuorumSigAt", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.VerifyQuorumSig(
		context.Background(),
		addr,
		symbiotic.Epoch(0),
		[]byte("test message"),
		symbiotic.KeyTag(1),
		big.NewInt(100),
		[]byte("proof"),
	)

	require.Error(t, err)
	assert.False(t, result)
}
