package evm

import (
	"context"
	"math/big"
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

func TestSetGenesis_NoSettlementContract_ReturnsError(t *testing.T) {
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

	header := symbiotic.ValidatorSetHeader{
		Version:            1,
		RequiredKeyTag:     symbiotic.KeyTag(1),
		Epoch:              symbiotic.Epoch(10),
		CaptureTimestamp:   symbiotic.Timestamp(1000),
		QuorumThreshold:    symbiotic.VotingPower{Int: big.NewInt(100)},
		TotalVotingPower:   symbiotic.VotingPower{Int: big.NewInt(1000)},
		ValidatorsSszMRoot: common.HexToHash("0xabcd"),
	}

	extraData := []symbiotic.ExtraData{
		{
			Key:   common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
			Value: common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
		},
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	// Metrics are called with "error" because settlement contract doesn't exist
	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, extraData)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_KeyProviderFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	header := symbiotic.ValidatorSetHeader{
		Version: 1,
		Epoch:   symbiotic.Epoch(10),
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

	result, err := client.SetGenesis(context.Background(), addr, header, nil)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_InvalidECDSAKey_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	chainID := uint64(1)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	header := symbiotic.ValidatorSetHeader{
		Version: 1,
		Epoch:   symbiotic.Epoch(10),
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

	result, err := client.SetGenesis(context.Background(), addr, header, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length")
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_ContextTimeout_ReturnsError(t *testing.T) {
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

	header := symbiotic.ValidatorSetHeader{
		Version: 1,
		Epoch:   symbiotic.Epoch(10),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

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

	result, err := client.SetGenesis(ctx, addr, header, nil)

	require.Error(t, err)
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_InvalidChainID_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKeyProv := mocks.NewMockkeyProvider(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	header := symbiotic.ValidatorSetHeader{
		Version: 1,
		Epoch:   symbiotic.Epoch(10),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_EmptyExtraData_HandlesCorrectly(t *testing.T) {
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

	header := symbiotic.ValidatorSetHeader{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(1),
		Epoch:            symbiotic.Epoch(10),
		CaptureTimestamp: symbiotic.Timestamp(1000),
		QuorumThreshold:  symbiotic.VotingPower{Int: big.NewInt(100)},
		TotalVotingPower: symbiotic.VotingPower{Int: big.NewInt(1000)},
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, []symbiotic.ExtraData{})

	require.Error(t, err)
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_NilExtraData_HandlesCorrectly(t *testing.T) {
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

	header := symbiotic.ValidatorSetHeader{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(1),
		Epoch:            symbiotic.Epoch(10),
		CaptureTimestamp: symbiotic.Timestamp(1000),
		QuorumThreshold:  symbiotic.VotingPower{Int: big.NewInt(100)},
		TotalVotingPower: symbiotic.VotingPower{Int: big.NewInt(1000)},
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]*ethclient.Client),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, nil)

	require.Error(t, err)
	assert.Empty(t, result.TxHash)
}
