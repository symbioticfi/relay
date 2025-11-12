package evm

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
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
		conns: map[uint64]conn{
			1:  mockConn1,
			2:  mockConn2,
			56: mockConn3,
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
		conns:   make(map[uint64]conn),
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
		conns: map[uint64]conn{
			1: mockConn,
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
		conns:   make(map[uint64]conn),
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
		conns:   make(map[uint64]conn),
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
		conns:   make(map[uint64]conn),
		metrics: nil,
	}

	startTime := time.Now()

	require.NotPanics(t, func() {
		client.observeMetrics("TestMethod", 1, nil, startTime)
	})
}
