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

	deploymentData := loadDeploymentData(t)
	t.Logf("Deployment data loaded: chainId=%d, existingOperators=%d", deploymentData.Driver.ChainId, deploymentData.Env.Operators)

	t.Log("Creating extra operator with keys...")
	opData := createExtraOperator(t)
	t.Logf("Extra operator created: address=%s", opData.address.Hex())

	t.Log("Creating EVM client for extra operator...")
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, opData.privateKey)
	t.Logf("Registering operator in OperatorRegistry at %s...", deploymentData.MainChain.Addresses.OperatorRegistry)
	_, err := opEVMClient.RegisterOperator(t.Context(), symbiotic.CrossChainAddress{
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

	t.Log("Creating operator's test EVM client...")
	opTestEVM := createTestEVM(t, settlementChains[0], opData.privateKey)
	t.Logf("Operator opting into network at %s...", deploymentData.MainChain.Addresses.Network)
	_, err = opTestEVM.OptIn(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.OperatorNetworkOptInService), common.HexToAddress(deploymentData.MainChain.Addresses.Network))
	if err != nil && !strings.Contains(err.Error(), "custom error 0xdcdeaba3") { // already opted in
		require.NoError(t, err)
	}
	t.Log("Operator opted into network")

	t.Log("Registering extra operator in VotingPowerProvider...")
	registerExtraOperator(t, opEVMClient, opData.address)
	t.Log("Extra operator registered in VotingPowerProvider")

	t.Log("Initializing vault for operator...")
	initVault(t, opTestEVM, funderTestEVM, opData.address)
	t.Log("Vault initialized for operator")

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, opData.privateKey.Bytes())
	require.NoError(t, err)

	t.Logf("Waiting for validator set to include %d validators (existing %d + 1 new)...", deploymentData.Env.Operators+1, deploymentData.Env.Operators)
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		valset, err := getValset(t, opEVMClient)
		if err != nil {
			return err
		}

		t.Logf("Current epoch %d has %d validators (expecting %d)", valset.Epoch, len(valset.Validators), deploymentData.Env.Operators+1)
		if int64(len(valset.Validators)) != deploymentData.Env.Operators+1 {
			return errors.Errorf("expected %d validators, got %d", deploymentData.Env.Operators+1, len(valset.Validators))
		}

		validator, found := valset.FindValidatorByKey(symbiotic.KeyTag(15), privateKey.PublicKey().OnChain())
		require.Truef(t, found, "extra operator's BLS key not found in validator set")
		t.Logf("Found extra operator's key in validator set")
		require.Len(t, validator.Keys, 3)

		return nil
	}))
	t.Log("Validator set now includes the extra operator")

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	require.NoError(t, err)
	t.Logf("Current epoch: %d", currentEpoch)

	t.Logf("Getting gRPC client for extra operator (index %d)...", deploymentData.Env.Operators)
	extraClient := getGRPCClient(t, int(deploymentData.Env.Operators))
	t.Logf("Fetching signatures for epoch %d...", currentEpoch)

	err = waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		resp, err := extraClient.GetSignaturesByEpoch(t.Context(),
			&apiv1.GetSignaturesByEpochRequest{
				Epoch: uint64(currentEpoch),
			},
		)
		require.NoError(t, err)
		t.Logf("Received %d signatures for epoch %d", len(resp.GetSignatures()), currentEpoch)

		if len(resp.GetSignatures()) != int(deploymentData.Env.Operators+1) {
			return errors.Errorf("expected %d signatures, got %d", deploymentData.Env.Operators+1, len(resp.GetSignatures()))
		}

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

	t.Logf("[registerExtraOperator] Registering operator %s in VotingPowerProvider at %s...", opData.address.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	_, err = opEVMClient.UnregisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)
	t.Log("[registerExtraOperator] RegisterOperatorVotingPowerProvider transaction sent")

	t.Logf("Waiting for validator set to exclude extra validator")
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute*3, func() error {
		valset, err := getValset(t, opEVMClient)
		if err != nil {
			return err
		}
		t.Logf("Current epoch %d has %d validators (expecting %d)", valset.Epoch, len(valset.Validators), deploymentData.Env.Operators)
		if int64(len(valset.Validators)) != deploymentData.Env.Operators {
			return errors.Errorf("expected %d validators, got %d", deploymentData.Env.Operators, len(valset.Validators))
		}

		return nil
	}))
	t.Log("Validator set now excludes the extra operator")
}

func getValset(t *testing.T, opEVMClient *evm.Client) (symbiotic.ValidatorSet, error) {
	t.Helper()
	deriver, err := valsetDeriver.NewDeriver(opEVMClient)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	currentEpoch, err := opEVMClient.GetCurrentEpoch(t.Context())
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	captureTimestamp, err := opEVMClient.GetEpochStart(t.Context(), currentEpoch)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	currentConfig, err := opEVMClient.GetConfig(t.Context(), captureTimestamp, currentEpoch)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	valset, err := deriver.GetValidatorSet(t.Context(), currentEpoch, currentConfig)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	return valset, nil
}

func registerExtraOperator(t *testing.T, opEVMClient *evm.Client, opAddress common.Address) {
	t.Helper()
	deploymentData := loadDeploymentData(t)

	t.Logf("[registerExtraOperator] Registering operator %s in VotingPowerProvider at %s...", opAddress.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	_, err := opEVMClient.RegisterOperatorVotingPowerProvider(t.Context(), symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
	})
	require.NoError(t, err)
	t.Log("[registerExtraOperator] RegisterOperatorVotingPowerProvider transaction sent")

	t.Log("[registerExtraOperator] Waiting for operator registration to be confirmed on-chain...")
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute, func() error {
		registered, err := opEVMClient.IsOperatorRegistered(t.Context(), symbiotic.CrossChainAddress{
			ChainId: deploymentData.Driver.ChainId,
			Address: common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider),
		}, symbiotic.CrossChainAddress{ChainId: deploymentData.Driver.ChainId, Address: opAddress})
		require.NoError(t, err)
		if !registered {
			t.Logf("[registerExtraOperator] Operator %s not registered yet, retrying...", opAddress.Hex())
			return errors.Errorf("operator %s not registered yet", opAddress.Hex())
		}
		t.Logf("[registerExtraOperator] Operator %s registration confirmed", opAddress.Hex())
		return nil
	}))
}

func initVault(t *testing.T, opTestEVM *testEth.Client, funderTestEVM *testEth.Client, address common.Address) {
	t.Helper()
	deploymentData := loadDeploymentData(t)
	stakingAmount := big.NewInt(1e5)

	t.Logf("[initVault] Getting auto-deploy vault for operator %s from VotingPowerProvider %s...", address.Hex(), deploymentData.MainChain.Addresses.VotingPowerProvider)
	vaultAddress, err := funderTestEVM.GetAutoDeployVault(t.Context(), common.HexToAddress(deploymentData.MainChain.Addresses.VotingPowerProvider), address)
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

	t.Logf("[createExtraOperator] Creating extra operator (existing operators: %d)", deploymentData.Env.Operators)

	// Matches generate_network.sh: BASE_PRIVATE_KEY + operators
	// BASE_PRIVATE_KEY = 1000000000000000000
	baseKey := big.NewInt(1000000000000000000)
	operators := big.NewInt(deploymentData.Env.Operators)
	extraKeyInt := new(big.Int).Add(baseKey, operators)
	extraSecondaryKeyInt := new(big.Int).Add(extraKeyInt, big.NewInt(10000))
	t.Logf("[createExtraOperator] Derived key: baseKey=%s + operators=%d = %s", baseKey.String(), deploymentData.Env.Operators, extraKeyInt.String())

	pkBytes := make([]byte, 32)
	extraKeyInt.FillBytes(pkBytes)
	pkSecondaryBytes := make([]byte, 32)
	extraSecondaryKeyInt.FillBytes(pkSecondaryBytes)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, pkBytes)
	require.NoError(t, err)
	t.Log("[createExtraOperator] Created ECDSA private key")

	// Derive Ethereum address from private key
	ecdsaKey, err := ethCrypto.ToECDSA(privateKey.Bytes())
	require.NoError(t, err)
	operatorAddress := ethCrypto.PubkeyToAddress(ecdsaKey.PublicKey)
	t.Logf("[createExtraOperator] Operator address: %s", operatorAddress.Hex())

	t.Logf("[createExtraOperator] Funding operator with 1 ETH on chain %d...", deploymentData.Driver.ChainId)
	_, err = fundOperator(t.Context(), getFunderPrivateKey(t), settlementChains[0], symbiotic.CrossChainAddress{
		ChainId: deploymentData.Driver.ChainId,
		Address: operatorAddress,
	}, big.NewInt(1e18))
	require.NoError(t, err)
	t.Log("[createExtraOperator] Operator funded")

	blsPrivateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkBytes)
	require.NoError(t, err)
	blsPrivateKeySecondary, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pkSecondaryBytes)
	require.NoError(t, err)
	t.Logf("[createExtraOperator] Created BLS keys - primary pubKey: %x", blsPrivateKey.PublicKey().OnChain())

	t.Log("[createExtraOperator] Creating EVM client and key registerer...")
	opEVMClient := createEVMClientWithEVMKey(t, deploymentData, privateKey)
	registerer, err := key_registerer.NewRegisterer(key_registerer.Config{EVMClient: opEVMClient})
	require.NoError(t, err)

	t.Logf("[createExtraOperator] Registering BLS key with keyTag=15 for operator %s...", operatorAddress.Hex())
	_, err = registerer.Register(t.Context(), blsPrivateKey, symbiotic.KeyTag(15), operatorAddress)
	require.NoError(t, err)
	t.Log("[createExtraOperator] BLS key (keyTag=15) registered")

	t.Logf("[createExtraOperator] Registering ECDSA key with keyTag=16 for operator %s...", operatorAddress.Hex())
	_, err = registerer.Register(t.Context(), privateKey, symbiotic.KeyTag(16), operatorAddress)
	require.NoError(t, err)
	t.Log("[createExtraOperator] ECDSA key (keyTag=16) registered")

	t.Logf("[createExtraOperator] Registering secondary BLS key with keyTag=11 for operator %s...", operatorAddress.Hex())
	_, err = registerer.Register(t.Context(), blsPrivateKeySecondary, symbiotic.KeyTag(11), operatorAddress)
	require.NoError(t, err)
	t.Log("[createExtraOperator] Secondary BLS key (keyTag=11) registered")

	t.Logf("[createExtraOperator] Extra operator created successfully: address=%s", operatorAddress.Hex())
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
