package evm

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// RegisterOperator tests

func TestRegisterOperator_KeyProviderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	expectedErr := errors.New("key provider error")
	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(nil, expectedErr)

	// Note: Metrics are not called because the error occurs before the defer is set up

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperator(context.Background(), addr)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result.TxHash)
}

func TestRegisterOperator_InvalidECDSAKey_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: nil}, nil)

	// Note: Metrics are not called because the error occurs before the defer is set up

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperator(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Empty(t, result.TxHash)
}

func TestRegisterOperator_ContextTimeout_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("RegisterOperator", chainID, "error", gomock.Any())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		cfg: Config{
			RequestTimeout: 1 * time.Millisecond,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperator(ctx, addr)

	require.Error(t, err)
	assert.Empty(t, result.TxHash)
}

func TestRegisterOperator_NoOperatorRegistry_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("RegisterOperator", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperator(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}

// RegisterKey tests

func TestRegisterKey_KeyProviderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	expectedErr := errors.New("key provider error")
	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterKey(context.Background(), addr, symbiotic.KeyTag(1), nil, nil, nil)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result.TxHash)
}

func TestRegisterKey_InvalidECDSAKey_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: nil}, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterKey(context.Background(), addr, symbiotic.KeyTag(1), nil, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Empty(t, result.TxHash)
}

func TestRegisterKey_NoKeyRegistry_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetKey", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterKey(context.Background(), addr, symbiotic.KeyTag(1), []byte("key"), []byte("signature"), []byte("extra"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}

func TestRegisterKey_EmptyExtraData_HandlesCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetKey", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterKey(context.Background(), addr, symbiotic.KeyTag(1), []byte("key"), []byte("signature"), []byte{})

	require.Error(t, err)
	assert.Empty(t, result.TxHash)
}

// RegisterOperatorVotingPowerProvider tests

func TestRegisterOperatorVotingPowerProvider_KeyProviderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	expectedErr := errors.New("key provider error")
	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result.TxHash)
}

func TestRegisterOperatorVotingPowerProvider_InvalidECDSAKey_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: nil}, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Empty(t, result.TxHash)
}

func TestRegisterOperatorVotingPowerProvider_NoVotingPowerProvider_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("RegisterOperatorVotingPowerProvider", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.RegisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get voting power provider contract")
	assert.Empty(t, result.TxHash)
}

// UnregisterOperatorVotingPowerProvider tests

func TestUnregisterOperatorVotingPowerProvider_KeyProviderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	expectedErr := errors.New("key provider error")
	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(nil, expectedErr)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.UnregisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result.TxHash)
}

func TestUnregisterOperatorVotingPowerProvider_InvalidECDSAKey_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: nil}, nil)

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.UnregisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Empty(t, result.TxHash)
}

func TestUnregisterOperatorVotingPowerProvider_NoVotingPowerProvider_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("UnregisterOperatorVotingPowerProvider", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.UnregisterOperatorVotingPowerProvider(context.Background(), addr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get voting power provider contract")
	assert.Empty(t, result.TxHash)
}
