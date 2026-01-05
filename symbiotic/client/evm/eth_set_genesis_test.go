package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type mockPrivateKey struct {
	key *ecdsa.PrivateKey
}

func (m *mockPrivateKey) Bytes() []byte {
	return crypto.FromECDSA(m.key)
}

func (m *mockPrivateKey) Sign(msg []byte) (symbiotic.RawSignature, symbiotic.MessageHash, error) {
	return nil, nil, nil
}

func (m *mockPrivateKey) PublicKey() symbiotic.PublicKey {
	return nil
}

func TestSetGenesis_NoSettlementContract_ReturnsError(t *testing.T) {
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

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]conn),
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

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   map[uint64]conn{1: nil},
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

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   map[uint64]conn{1: nil},
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
		Version:            1,
		RequiredKeyTag:     symbiotic.KeyTag(1),
		Epoch:              symbiotic.Epoch(10),
		CaptureTimestamp:   symbiotic.Timestamp(1000),
		QuorumThreshold:    symbiotic.VotingPower{Int: big.NewInt(100)},
		TotalVotingPower:   symbiotic.VotingPower{Int: big.NewInt(1000)},
		ValidatorsSszMRoot: common.HexToHash("0xabcd"),
	}

	mockKeyProv.EXPECT().
		GetPrivateKeyByNamespaceTypeId(gomock.Any(), symbiotic.KeyTypeEcdsaSecp256k1, int(chainID)).
		Return(&mockPrivateKey{key: privateKey}, nil)

	mockMetrics.EXPECT().
		ObserveEVMMethodCall("SetGenesis", chainID, "error", gomock.Any())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	mockconn := mocks.NewMockconn(ctrl)
	mockconn.EXPECT().HeaderByNumber(gomock.Any(), gomock.Any()).Return(nil, context.DeadlineExceeded)
	client := &Client{
		cfg: Config{
			RequestTimeout: 1 * time.Millisecond,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:         map[uint64]conn{1: mockconn},
		driverChainID: 1,
		metrics:       mockMetrics,
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

	chainID := uint64(999)
	addr := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	header := symbiotic.ValidatorSetHeader{
		Version: 1,
		Epoch:   symbiotic.Epoch(10),
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]conn),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}

func TestSetGenesis_PartialHappyPath_ValidatesDataPreparation(t *testing.T) {
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
		Version:            1,
		RequiredKeyTag:     symbiotic.KeyTag(1),
		Epoch:              symbiotic.Epoch(0),
		CaptureTimestamp:   symbiotic.Timestamp(1000),
		QuorumThreshold:    symbiotic.VotingPower{Int: big.NewInt(100)},
		TotalVotingPower:   symbiotic.VotingPower{Int: big.NewInt(1000)},
		ValidatorsSszMRoot: common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
	}

	extraData := []symbiotic.ExtraData{
		{
			Key:   common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
			Value: common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
		},
	}

	client := &Client{
		cfg: Config{
			RequestTimeout: 5 * time.Second,
			KeyProvider:    mockKeyProv,
			Metrics:        mockMetrics,
		},
		conns:   make(map[uint64]conn),
		metrics: mockMetrics,
	}

	result, err := client.SetGenesis(context.Background(), addr, header, extraData)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get settlement contract")
	assert.Empty(t, result.TxHash)
}
