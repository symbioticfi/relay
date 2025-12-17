package tests

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	testEth "github.com/symbioticfi/relay/e2e/tests/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

func TestAddAndRemoveOperator(t *testing.T) {
	deploymentData := loadDeploymentData(t)

	opData := createOperator(t)

	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)
	_, err := opEVMClient.RegisterOperator(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.OperatorRegistry),
	})
	require.NoError(t, err)

	funderTestEVM := createTestEVM(t, settlementChains[0], getFunderPrivateKey(t))
	stakingAmount := big.NewInt(1e5)
	_, err = funderTestEVM.TransferMockToken(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), opData.address, stakingAmount)
	require.NoError(t, err)

	opTestEVM := createTestEVM(t, settlementChains[0], opData.privateKey)
	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorNetworkOptInService), common.HexToAddress(deploymentData.MainChain.Addresses.Network))
	require.NoError(t, err)

	_, err = opEVMClient.RegisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)

	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute, func() error {
		registered, err := opEVMClient.IsOperatorRegistered(t.Context(), symbiotic.CrossChainAddress{
			ChainId: deploymentData.Driver.ChainId,
			Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
		}, symbiotic.CrossChainAddress{ChainId: deploymentData.Driver.ChainId, Address: opData.address})
		require.NoError(t, err)
		if !registered {
			return errors.Errorf("operator %s not registered yet", opData.address.Hex())
		}

		return nil
	}))

	vaultAddress, err := funderTestEVM.GetAutoDeployVault(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider), opData.address)
	require.NoError(t, err)

	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorVaultOptInService), vaultAddress)
	require.NoError(t, err)

	_, err = funderTestEVM.ApproveMockToken(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), vaultAddress, stakingAmount)
	require.NoError(t, err)

	_, err = funderTestEVM.VaultDeposit(t.Context(), vaultAddress, common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), stakingAmount)
	require.NoError(t, err)

	time.Sleep(time.Minute)

	deriver, err := valsetDeriver.NewDeriver(opEVMClient)
	require.NoError(t, err)

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	require.NoError(t, err, "Failed to get current epoch")
	t.Logf("Current epoch: %d", currentEpoch)

	captureTimestamp, err := opEVMClient.GetEpochStart(t.Context(), currentEpoch)
	require.NoError(t, err, "Failed to get epoch start timestamp")

	currentConfig, err := opEVMClient.GetConfig(t.Context(), captureTimestamp, currentEpoch)
	require.NoError(t, err, "Failed to get network config")

	valset, err := deriver.GetValidatorSet(t.Context(), currentEpoch, currentConfig)
	require.NoError(t, err)
	fmt.Println(valset) // TODO remove
}

func createTestEVM(t *testing.T, chainURL string, privateKey crypto.PrivateKey) *testEth.Client {
	t.Helper()
	return testEth.NewClient(t, testEth.Config{
		ChainURL:   chainURL,
		PrivateKey: privateKey,
	})
}

type operatorData struct {
	privateKey crypto.PrivateKey
	address    common.Address
}

func createOperator(t *testing.T) operatorData {
	t.Helper()
	deploymentData := loadDeploymentData(t)

	pkBytes := make([]byte, 32)
	_, err := rand.Read(pkBytes)
	require.NoError(t, err)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, pkBytes)
	require.NoError(t, err)

	// Derive Ethereum address from private key
	ecdsaKey, err := ethCrypto.ToECDSA(privateKey.Bytes())
	require.NoError(t, err)
	operatorAddress := ethCrypto.PubkeyToAddress(ecdsaKey.PublicKey)

	_, err = fundOperator(t.Context(), getFunderPrivateKey(t), settlementChains[0], symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: operatorAddress,
	}, big.NewInt(1e18))
	require.NoError(t, err)

	return operatorData{
		privateKey: privateKey,
		address:    operatorAddress,
	}
}

func fundOperator(
	ctx context.Context,
	pk crypto.PrivateKey,
	chainURL string,
	opAddr symbiotic.CrossChainAddress,
	amountWei *big.Int, // Amount in wei (e.g., 1 ether = 1e18 wei)
) (symbiotic.TxResult, error) {
	ecdsaKey, err := ethCrypto.ToECDSA(pk.Bytes())
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to convert to ECDSA key: %w", err)
	}

	client, err := ethclient.DialContext(ctx, chainURL)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to connect to Ethereum client: %w", err)
	}

	fromAddress := ethCrypto.PubkeyToAddress(ecdsaKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get gas price: %w", err)
	}

	chainIDBig := new(big.Int).SetUint64(opAddr.ChainId)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &opAddr.Address,
		Value:    amountWei,
		Gas:      21000, // Standard gas limit for ETH transfer
		GasPrice: gasPrice,
		Data:     nil,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainIDBig), ecdsaKey)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to sign transaction for operator %s: %w", opAddr.Address.Hex(), err)
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to send transaction for operator %s: %w", opAddr.Address.Hex(), err)
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, client, signedTx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to wait for tx mining for operator %s: %w", opAddr.Address.Hex(), err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return symbiotic.TxResult{}, errors.Errorf("transaction reverted on chain for operator %s", opAddr.Address.Hex())
	}

	return symbiotic.TxResult{
		TxHash:            receipt.TxHash,
		GasUsed:           receipt.GasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
	}, nil
}
