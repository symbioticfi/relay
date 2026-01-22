package evm

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
)

func TestMulticallExists_NoConnection_ReturnsError(t *testing.T) {
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

	exists, err := client.multicallExists(context.Background(), 999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.False(t, exists)
}

func TestMulticallExists_CodeAtFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)

	expectedErr := errors.New("rpc error")
	mockConn.EXPECT().
		CodeAt(gomock.Any(), common.HexToAddress(Multicall3), gomock.Any()).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   map[uint64]clientWithInfo{chainID: {conn: mockConn}},
		metrics: mockMetrics,
	}

	exists, err := client.multicallExists(context.Background(), chainID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get Multicall3 code")
	assert.False(t, exists)
}

func TestMulticallExists_CodeExists_ReturnsTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)

	mockConn.EXPECT().
		CodeAt(gomock.Any(), common.HexToAddress(Multicall3), gomock.Any()).
		Return([]byte{0x60, 0x80}, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   map[uint64]clientWithInfo{chainID: {conn: mockConn}},
		metrics: mockMetrics,
	}

	exists, err := client.multicallExists(context.Background(), chainID)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMulticallExists_NoCode_ReturnsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := mocks.NewMockconn(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)

	mockConn.EXPECT().
		CodeAt(gomock.Any(), common.HexToAddress(Multicall3), gomock.Any()).
		Return([]byte{}, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   map[uint64]clientWithInfo{chainID: {conn: mockConn}},
		metrics: mockMetrics,
	}

	exists, err := client.multicallExists(context.Background(), chainID)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMulticall_NoConnection_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(999)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("Multicall", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]clientWithInfo),
		metrics: mockMetrics,
	}

	calls := []Call{
		{
			Target:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			CallData:     []byte("test"),
			AllowFailure: false,
		},
	}

	result, err := client.multicall(context.Background(), chainID, calls)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection for chain ID")
	assert.Nil(t, result)
}
