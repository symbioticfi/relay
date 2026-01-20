package tests

import (
	"bytes"
	"context"
	"math/big"
	"strings"
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

func TestAddAndRemoveOperator(t *testing.T) {
	t.Log("=== Starting TestAddAndRemoveOperator ===")

	deployData := loadDeploymentData(t)
	t.Logf("Deployment data loaded: chainId=%d, existingOperators=%d", deployData.Driver.ChainId, deployData.Env.Operators)

	extraData := createExtraOperator(t)
	someOperator := newOperatorData(t, 1)

	unregisterOperator(t, someOperator)
	registerOperator(t, extraData)

	initVault(t, extraData)

	waitOperatorIncludedIntoValset(t, extraData, int(deployData.Env.Operators))

	checkOperatorProducesSignatures(t, extraData)

	unregisterOperator(t, extraData)
	registerOperator(t, someOperator)

	waitForNextCommitment(t, createEVMClientWithEVMKey(t, deployData, extraData.privateKey))
}

func checkOperatorProducesSignatures(t *testing.T, opData operatorData) {
	t.Helper()
	deploymentData := loadDeploymentData(t)
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	require.NoError(t, err)
	t.Logf("Current epoch: %d", currentEpoch)

	t.Logf("Getting gRPC client for extra operator (index %d)...", opData.number)
	extraClient := getGRPCClient(t, opData.number)
	t.Logf("Fetching signatures for epoch %d...", currentEpoch)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, opData.privateKey.Bytes())
	require.NoError(t, err)

	err = waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		resp, err := extraClient.GetSignaturesByEpoch(t.Context(),
			&apiv1.GetSignaturesByEpochRequest{
				Epoch: uint64(currentEpoch),
			},
		)
		require.NoError(t, err)
		t.Logf("Received %d signatures for epoch %d", len(resp.GetSignatures()), currentEpoch)

		expectedPubKey := privateKey.PublicKey().Raw()
		t.Logf("Looking for signature from extra operator with public key: %x", expectedPubKey)

		for i, signature := range resp.GetSignatures() {
			t.Logf("Signature %d: pubKey=%x, requestId=%s", i, signature.GetPublicKey(), signature.GetRequestId())
			if bytes.Equal(signature.GetPublicKey(), expectedPubKey) {
				t.Log("=== SUCCESS: Found signature from extra operator ===")
				return nil
			}
		}

		return errors.New("did not find signature from extra operator")
	})
	require.NoError(t, err)
}

func waitOperatorIncludedIntoValset(t *testing.T, opData operatorData, expectedCount int) {
	t.Helper()

	deploymentData := loadDeploymentData(t)
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, opData.privateKey.Bytes())
	require.NoError(t, err)

	t.Logf("Waiting for validator set to include %d validators", expectedCount)
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		valset := getValset(t, opEVMClient)

		t.Logf("Current epoch %d has %d validators (expecting %d)", valset.Epoch, len(valset.Validators), expectedCount)
		if len(valset.Validators) != expectedCount {
			return errors.Errorf("expected %d validators, got %d", expectedCount, len(valset.Validators))
		}

		validator, found := valset.FindValidatorByKey(symbiotic.KeyTag(15), privateKey.PublicKey().OnChain())
		if !found {
			return errors.New("extra operator's BLS key not found in validator set")
		}
		t.Logf("Found extra operator's key in validator set")
		require.Len(t, validator.Keys, 3)

		return nil
	}))
	t.Log("Validator set now includes the extra operator")
}

func getValset(t *testing.T, opEVMClient *evm.Client) symbiotic.ValidatorSet {
	t.Helper()
	deriver, err := valsetDeriver.NewDeriver(opEVMClient)
	require.NoError(t, err)

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	require.NoError(t, err)

	captureTimestamp, err := opEVMClient.GetEpochStart(t.Context(), currentEpoch)
	require.NoError(t, err)

	currentConfig, err := opEVMClient.GetConfig(t.Context(), captureTimestamp, currentEpoch)
	require.NoError(t, err)

	valset, err := deriver.GetValidatorSet(t.Context(), currentEpoch, currentConfig)
	require.NoError(t, err)

	return valset
}

func unregisterOperator(t *testing.T, opData operatorData) {
	t.Helper()

	opEVMClient := createEVMClientWithEVMKey(t, loadDeploymentData(t), opData.privateKey)
	prevValset := getValset(t, opEVMClient)

	deploymentData := loadDeploymentData(t)
	t.Logf("Unregistering operator %s in VotingPowerProvider at %s...", opData.address.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	_, err := opEVMClient.UnregisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)
	t.Log("UnregisterOperatorVotingPowerProvider transaction sent")

	t.Logf("Waiting for validator set to exclude validator")
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		valset := getValset(t, opEVMClient)

		t.Logf("Current epoch %d has %d validators (expecting %d)", valset.Epoch, len(valset.Validators), len(prevValset.Validators)-1)
		if len(valset.Validators) != len(prevValset.Validators)-1 {
			return errors.Errorf("expected %d validators, got %d", len(prevValset.Validators)-1, len(valset.Validators))
		}

		return nil
	}))
	t.Log("Validator set now excludes the extra operator")
}

func registerOperator(t *testing.T, opData operatorData) {
	t.Helper()
	deploymentData := loadDeploymentData(t)
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)

	t.Log("Registering operator in VotingPowerProvider...")

	t.Logf("[registerOperator] Registering operator %s in VotingPowerProvider at %s...", opData.address.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	_, err := opEVMClient.RegisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)
	t.Log("[registerOperator] RegisterOperatorVotingPowerProvider transaction sent")

	t.Log("[registerOperator] Waiting for operator registration to be confirmed on-chain...")
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute, func() error {
		registered, err := opEVMClient.IsOperatorRegistered(t.Context(), symbiotic.CrossChainAddress{
			ChainId: deploymentData.Driver.ChainId,
			Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
		}, symbiotic.CrossChainAddress{ChainId: deploymentData.Driver.ChainId, Address: opData.address})
		require.NoError(t, err)

		if !registered {
			t.Logf("[registerOperator] Operator %s not registered yet, retrying...", opData.address.Hex())
			return errors.Errorf("operator %s not registered yet", opData.address.Hex())
		}
		t.Logf("[registerOperator] Operator %s registration confirmed", opData.address.Hex())
		return nil
	}))

	t.Log("Extra operator registered in VotingPowerProvider")

	opTestEVM := createTestEVM(t, settlementChains[0], opData.privateKey)
	t.Logf("Operator opting into network at %s...", deploymentData.MainChain.Addresses.Network)
	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorNetworkOptInService), common.HexToAddress(deploymentData.MainChain.Addresses.Network))
	if err != nil && !strings.Contains(err.Error(), "custom error 0xdcdeaba3") { // already opted in
		require.NoError(t, err)
	}
	t.Log("Operator opted into network")
}

func initVault(t *testing.T, opData operatorData) {
	t.Helper()
	t.Log("Initializing vault for operator...")

	deploymentData := loadDeploymentData(t)
	createTestEVM(t, settlementChains[0], opData.privateKey)
	stakingAmount := big.NewInt(1e5)

	funderTestEVM := createTestEVM(t, settlementChains[0], getFunderPrivateKey(t))
	opTestEVM := createTestEVM(t, settlementChains[0], opData.privateKey)

	t.Logf("[initVault] Getting auto-deploy vault for operator %s from VotingPowerProvider %s...", opData.address.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	vaultAddress, err := funderTestEVM.GetAutoDeployVault(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider), opData.address)
	require.NoError(t, err)
	t.Logf("[initVault] Vault address: %s", vaultAddress.Hex())

	t.Logf("[initVault] Operator opting into vault %s via OperatorVaultOptInService %s...", vaultAddress.Hex(), deploymentData.MainChain.Addresses.OperatorVaultOptInService)
	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorVaultOptInService), vaultAddress)
	if err != nil && !strings.Contains(err.Error(), "custom error 0xdcdeaba3") { // already opted in
		require.NoError(t, err)
	}
	t.Log("[initVault] Operator opted into vault")

	t.Logf("[initVault] Approving %s staking tokens for vault %s...", stakingAmount.String(), vaultAddress.Hex())
	_, err = funderTestEVM.ApproveMockToken(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), vaultAddress, stakingAmount)
	require.NoError(t, err)
	t.Log("[initVault] Token approval complete")

	t.Logf("[initVault] Depositing %s tokens into vault %s...", stakingAmount.String(), vaultAddress.Hex())
	_, err = funderTestEVM.VaultDeposit(t.Context(), vaultAddress, common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), stakingAmount)
	require.NoError(t, err)
	t.Log("[initVault] Vault deposit complete")

	t.Log("Vault initialized for operator")
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

func newOperatorData(t *testing.T, operatorNumber int64) operatorData {
	t.Helper()

	// Matches generate_network.sh: BASE_PRIVATE_KEY + operators
	// BASE_PRIVATE_KEY = 1000000000000000000
	baseKey := big.NewInt(1000000000000000000)
	operators := big.NewInt(operatorNumber)
	extraKeyInt := new(big.Int).Add(baseKey, operators)
	extraSecondaryKeyInt := new(big.Int).Add(extraKeyInt, big.NewInt(10000))
	t.Logf("[createExtraOperator] Derived key: baseKey=%s + operators=%d = %s", baseKey.String(), operatorNumber, extraKeyInt.String())

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

	blsPrivateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkBytes)
	require.NoError(t, err)
	blsPrivateKeySecondary, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkSecondaryBytes)
	require.NoError(t, err)
	t.Logf("[createExtraOperator] Created BLS keys - primary pubKey: %x", blsPrivateKey.PublicKey().OnChain())

	return operatorData{
		number:                 int(operatorNumber),
		privateKey:             privateKey,
		blsPrivateKey:          blsPrivateKey,
		blsPrivateKeySecondary: blsPrivateKeySecondary,
		address:                operatorAddress,
	}
}

type operatorData struct {
	number                 int
	privateKey             crypto.PrivateKey
	address                common.Address
	blsPrivateKey          crypto.PrivateKey
	blsPrivateKeySecondary crypto.PrivateKey
}

// createExtraOperator creates an operator using the extra relay's private key.
// The extra relay is not part of the validator set and uses a deterministic key
// based on the formula: BASE_PRIVATE_KEY + operators (matching generate_network.sh).
func createExtraOperator(t *testing.T) operatorData {
	t.Helper()
	t.Log("Creating extra operator with keys...")
	deploymentData := loadDeploymentData(t)

	t.Logf("[createExtraOperator] Creating extra operator (existing operators: %d)", deploymentData.Env.Operators)

	opData := newOperatorData(t, deploymentData.Env.Operators)

	t.Logf("[createExtraOperator] Funding operator with 1 ETH on chain %d...", deploymentData.Driver.ChainId)
	_, err := fundOperator(t.Context(), getFunderPrivateKey(t), settlementChains[0], symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: opData.address,
	}, big.NewInt(1e18))
	require.NoError(t, err)
	t.Log("[createExtraOperator] Operator funded")

	t.Log("[createExtraOperator] Creating EVM client and key registerer...")
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)
	registerer, err := key_registerer.NewRegisterer(key_registerer.Config{EVMClient: opEVMClient})
	require.NoError(t, err)

	t.Logf("[createExtraOperator] Registering BLS key with keyTag=15 for operator %s...", opData.address.Hex())
	_, err = registerer.Register(t.Context(), opData.blsPrivateKey, symbiotic.KeyTag(15), opData.address)
	require.NoError(t, err)
	t.Log("[createExtraOperator] BLS key (keyTag=15) registered")

	t.Logf("[createExtraOperator] Registering ECDSA key with keyTag=16 for operator %s...", opData.address.Hex())
	_, err = registerer.Register(t.Context(), opData.privateKey, symbiotic.KeyTag(16), opData.address)
	require.NoError(t, err)
	t.Log("[createExtraOperator] ECDSA key (keyTag=16) registered")

	t.Logf("[createExtraOperator] Registering secondary BLS key with keyTag=11 for operator %s...", opData.address.Hex())
	_, err = registerer.Register(t.Context(), opData.blsPrivateKeySecondary, symbiotic.KeyTag(11), opData.address)
	require.NoError(t, err)
	t.Log("[createExtraOperator] Secondary BLS key (keyTag=11) registered")

	t.Logf("Registering operator in OperatorRegistry at %s...", deploymentData.MainChain.Addresses.OperatorRegistry)
	_, err = opEVMClient.RegisterOperator(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.OperatorRegistry),
	})
	if err != nil && !strings.Contains(err.Error(), "error OperatorAlreadyRegistered()") {
		require.NoError(t, err)
	}
	t.Log("Operator registered in OperatorRegistry")

	t.Log("Creating funder EVM client...")
	funderTestEVM := createTestEVM(t, settlementChains[0], getFunderPrivateKey(t))
	stakingAmount := big.NewInt(1e5)
	t.Logf("Transferring %s staking tokens to operator %s...", stakingAmount.String(), opData.address.Hex())
	_, err = funderTestEVM.TransferMockToken(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.StakingToken), opData.address, stakingAmount)
	require.NoError(t, err)
	t.Log("Staking tokens transferred to operator")

	t.Logf("Extra operator created: address=%s", opData.address.Hex())

	return opData
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
