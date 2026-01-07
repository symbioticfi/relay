package tests

import (
	"bytes"
	"context"
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

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	testEth "github.com/symbioticfi/relay/e2e/tests/evm"
	key_registerer "github.com/symbioticfi/relay/internal/usecase/key-registerer"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

// 4. проверить, что он генерит подписи
// 5. стопнуть лишнего оператора
func TestAddAndRemoveOperator(t *testing.T) {
	deploymentData := loadDeploymentData(t)

	opData := createExtraOperator(t)

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

	registerExtraOperator(t, opEVMClient, opData.address)
	initVault(t, opTestEVM, funderTestEVM, opData.address)

	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute, func() error {
		deriver, err := valsetDeriver.NewDeriver(opEVMClient)
		if err != nil {
			return err
		}

		currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
		if err != nil {
			return err
		}

		captureTimestamp, err := opEVMClient.GetEpochStart(t.Context(), currentEpoch)
		if err != nil {
			return err
		}

		currentConfig, err := opEVMClient.GetConfig(t.Context(), captureTimestamp, currentEpoch)
		if err != nil {
			return err
		}

		valset, err := deriver.GetValidatorSet(t.Context(), currentEpoch, currentConfig)
		if err != nil {
			return err
		}
		if int64(len(valset.Validators)) != deploymentData.Env.Operators+1 {
			return errors.Errorf("expected %d validators, got %d", deploymentData.Env.Operators+1, len(valset.Validators))
		}

		return nil
	}))

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	require.NoError(t, err)

	time.Sleep(time.Minute)

	extraClient := getGRPCClient(t, int(deploymentData.Env.Operators))
	resp, err := extraClient.GetSignaturesByEpoch(t.Context(),
		&apiv1.GetSignaturesByEpochRequest{
			Epoch: uint64(currentEpoch),
		},
	)
	require.NoError(t, err)

	require.Len(t, resp.GetSignatures(), int(deploymentData.Env.Operators+1))
	for _, signature := range resp.GetSignatures() {
		pk, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, opData.privateKey.Bytes())
		require.NoError(t, err)
		if bytes.Compare(signature.GetPublicKey(), pk.PublicKey().OnChain()) == 0 {
			t.Log("found signature from extra operator")
			return
		}
	}

	t.Fatal("did not find signature from extra operator")
}

func registerExtraOperator(t *testing.T, opEVMClient *evm.Client, opAddress common.Address) {
	deploymentData := loadDeploymentData(t)

	_, err := opEVMClient.RegisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)

	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute, func() error {
		registered, err := opEVMClient.IsOperatorRegistered(t.Context(), symbiotic.CrossChainAddress{
			ChainId: deploymentData.Driver.ChainId,
			Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
		}, symbiotic.CrossChainAddress{ChainId: deploymentData.Driver.ChainId, Address: opAddress})
		require.NoError(t, err)
		if !registered {
			return errors.Errorf("operator %s not registered yet", opAddress.Hex())
		}

		return nil
	}))
}

func initVault(t *testing.T, opTestEVM *testEth.Client, funderTestEVM *testEth.Client, address common.Address) {
	deploymentData := loadDeploymentData(t)
	stakingAmount := big.NewInt(1e5)

	vaultAddress, err := funderTestEVM.GetAutoDeployVault(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider), address)
	require.NoError(t, err)

	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorVaultOptInService), vaultAddress)
	require.NoError(t, err)

	_, err = funderTestEVM.ApproveMockToken(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), vaultAddress, stakingAmount)
	require.NoError(t, err)

	_, err = funderTestEVM.VaultDeposit(t.Context(), vaultAddress, common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), stakingAmount)
	require.NoError(t, err)
}

func createTestEVM(t *testing.T, chainURL string, privateKey crypto.PrivateKey) *testEth.Client {
	t.Helper()
	client := testEth.NewClient(t, testEth.Config{
		ChainURL:   chainURL,
		PrivateKey: privateKey,
	})
	t.Cleanup(func() {
		client.Close()
	})
	return client
}

type operatorData struct {
	privateKey crypto.PrivateKey
	address    common.Address
}

// createExtraOperator creates an operator using the extra relay's private key.
// The extra relay is not part of the validator set and uses a deterministic key
// based on the formula: BASE_PRIVATE_KEY + operators (matching generate_network.sh).
func createExtraOperator(t *testing.T) operatorData {
	t.Helper()
	deploymentData := loadDeploymentData(t)

	// Matches generate_network.sh: BASE_PRIVATE_KEY + operators
	// BASE_PRIVATE_KEY = 1000000000000000000
	baseKey := big.NewInt(1000000000000000000)
	operators := big.NewInt(deploymentData.Env.Operators)
	extraKeyInt := new(big.Int).Add(baseKey, operators)
	extraSecondaryKeyInt := new(big.Int).Add(extraKeyInt, big.NewInt(10000))

	pkBytes := make([]byte, 32)
	extraKeyInt.FillBytes(pkBytes)
	pkSecondaryBytes := make([]byte, 32)
	extraSecondaryKeyInt.FillBytes(pkSecondaryBytes)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, pkBytes)
	require.NoError(t, err)

	// Derive Ethereum address from private key
	ecdsaKey, err := ethCrypto.ToECDSA(privateKey.Bytes())
	require.NoError(t, err)
	operatorAddress := ethCrypto.PubkeyToAddress(ecdsaKey.PublicKey)
	t.Logf("Extra operator address: %s", operatorAddress.Hex())

	_, err = fundOperator(t.Context(), getFunderPrivateKey(t), settlementChains[0], symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: operatorAddress,
	}, big.NewInt(1e18))
	require.NoError(t, err)

	blsPrivateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkBytes)
	require.NoError(t, err)
	blsPrivateKeySecondary, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkSecondaryBytes)
	require.NoError(t, err)

	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, privateKey)
	registerer, err := key_registerer.NewRegisterer(key_registerer.Config{EVMClient: opEVMClient})
	require.NoError(t, err)

	_, err = registerer.Register(t.Context(), blsPrivateKey, symbiotic.KeyTag(15), operatorAddress)
	require.NoError(t, err)

	_, err = registerer.Register(t.Context(), privateKey, symbiotic.KeyTag(16), operatorAddress)
	require.NoError(t, err)

	_, err = registerer.Register(t.Context(), blsPrivateKeySecondary, symbiotic.KeyTag(11), operatorAddress)
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
