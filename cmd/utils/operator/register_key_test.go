package operator

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	evmmocks "github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestRegisterKeyUsesKeysProviderChainSecret(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	relayKeyBytes := make([]byte, 32)
	relayKeyBytes[len(relayKeyBytes)-1] = 0x66

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, relayKeyBytes)
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEVMClient := evmmocks.NewMockIEvmClient(ctrl)
	mockEVMClient.EXPECT().GetCurrentEpoch(gomock.Any()).Return(symbiotic.Epoch(7), nil)
	mockEVMClient.EXPECT().GetEpochStart(gomock.Any(), symbiotic.Epoch(7)).Return(symbiotic.Timestamp(123), nil)
	mockEVMClient.EXPECT().GetConfig(gomock.Any(), symbiotic.Timestamp(123), symbiotic.Epoch(7)).Return(symbiotic.NetworkConfig{
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: 137,
			Address: common.HexToAddress("0x0000000000000000000000000000000000000137"),
		},
	}, nil)
	mockEVMClient.EXPECT().GetChains().Return([]uint64{1, 137})

	builder := &capturingRegisterKeyBuilder{}
	restore := stubRegisterKeyCommandDependencies(t, mockEVMClient, builder)
	defer restore()

	chainOneSecret := "1111111111111111111111111111111111111111111111111111111111111111"
	keysProviderSecret := "2222222222222222222222222222222222222222222222222222222222222222"

	output, err := runOperatorCommand(t,
		"register-key",
		"--key-tag", keyTagArg(keyTag),
		"--path", keystorePath,
		"--password", testKeystorePassword,
		"--secret-keys", "1:"+chainOneSecret+",137:"+keysProviderSecret,
	)
	require.NoError(t, err, output)
	require.True(t, builder.registerCalled)

	expectedOperator, err := operatorAddressFromSecret(keysProviderSecret)
	require.NoError(t, err)
	require.Equal(t, expectedOperator, builder.operatorAddress)
}

func TestRegisterKeyFailsWhenKeysProviderChainIsNotConfigured(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEVMClient := evmmocks.NewMockIEvmClient(ctrl)
	mockEVMClient.EXPECT().GetCurrentEpoch(gomock.Any()).Return(symbiotic.Epoch(7), nil)
	mockEVMClient.EXPECT().GetEpochStart(gomock.Any(), symbiotic.Epoch(7)).Return(symbiotic.Timestamp(123), nil)
	mockEVMClient.EXPECT().GetConfig(gomock.Any(), symbiotic.Timestamp(123), symbiotic.Epoch(7)).Return(symbiotic.NetworkConfig{
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: 137,
			Address: common.HexToAddress("0x0000000000000000000000000000000000000137"),
		},
	}, nil)
	mockEVMClient.EXPECT().GetChains().Return([]uint64{1})

	builder := &capturingRegisterKeyBuilder{}
	restore := stubRegisterKeyCommandDependencies(t, mockEVMClient, builder)
	defer restore()

	output, err := runOperatorCommand(t,
		"register-key",
		"--key-tag", keyTagArg(keyTag),
	)
	require.EqualError(t, err, "keys provider chain 137 is not configured")
	require.Contains(t, output, "keys provider chain 137 is not configured")
	require.False(t, builder.registerCalled)
}

func stubRegisterKeyCommandDependencies(t *testing.T, evmClient *evmmocks.MockIEvmClient, builder registerKeyBuilder) func() {
	t.Helper()

	originalClientFactory := newRegisterKeyEVMClient
	originalBuilderFactory := newRegisterKeyBuilder
	originalMetricsFactory := newRegisterKeyMetrics

	newRegisterKeyMetrics = func() *metrics.Metrics {
		return metrics.New(metrics.Config{Registerer: prometheus.NewRegistry()})
	}
	newRegisterKeyEVMClient = func(_ context.Context, _ evm.Config) (evm.IEvmClient, error) {
		return evmClient, nil
	}
	newRegisterKeyBuilder = func(evm.IEvmClient) (registerKeyBuilder, error) {
		return builder, nil
	}

	return func() {
		newRegisterKeyMetrics = originalMetricsFactory
		newRegisterKeyEVMClient = originalClientFactory
		newRegisterKeyBuilder = originalBuilderFactory
	}
}

type capturingRegisterKeyBuilder struct {
	registerCalled  bool
	operatorAddress common.Address
}

func (c *capturingRegisterKeyBuilder) Register(
	_ context.Context,
	_ symbioticCrypto.PrivateKey,
	_ symbiotic.KeyTag,
	operatorAddress common.Address,
) (symbiotic.TxResult, error) {
	c.registerCalled = true
	c.operatorAddress = operatorAddress
	return symbiotic.TxResult{}, nil
}

func operatorAddressFromSecret(secret string) (common.Address, error) {
	privateKey, err := gethcrypto.HexToECDSA(secret)
	if err != nil {
		return common.Address{}, err
	}
	return gethcrypto.PubkeyToAddress(privateKey.PublicKey), nil
}
