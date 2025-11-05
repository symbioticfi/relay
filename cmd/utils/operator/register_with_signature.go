package operator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var registerOperatorWithSignatureCmd = &cobra.Command{
	Use:   "register-operator-with-signature",
	Short: "Generate EIP-712 signature for operator registration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())

		evmClient, err := evm.NewEvmClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: globalFlags.DriverChainId,
				Address: common.HexToAddress(globalFlags.DriverAddress),
			},
			RequestTimeout: 5 * time.Second,
			Metrics:        metrics.New(metrics.Config{}),
		})
		if err != nil {
			return err
		}

		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
		secret, ok := registerOperatorWithSignatureFlags.Secrets.Secrets[globalFlags.DriverChainId]
		if !ok {
			secret, _ = privateKeyInput.Show("Enter operator private key for chain with ID: " + strconv.Itoa(int(globalFlags.DriverChainId)))
		}

		ecdsaPk, err := crypto.HexToECDSA(secret)
		if err != nil {
			return errors.Errorf("failed to parse private key: %w", err)
		}
		operator := crypto.PubkeyToAddress(ecdsaPk.PublicKey)

		currentOnchainEpoch, err := evmClient.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		captureTimestamp, err := evmClient.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		// Use the first VotingPowerProvider from network config
		if len(networkConfig.VotingPowerProviders) == 0 {
			return errors.New("no voting power providers found in network config")
		}
		votingPowerProvider := networkConfig.VotingPowerProviders[0]

		eip712Domain, err := evmClient.GetVotingPowerProviderEip712Domain(ctx, votingPowerProvider)
		if err != nil {
			return errors.Errorf("failed to get eip712 domain: %w", err)
		}

		nonce, err := evmClient.GetOperatorNonce(ctx, votingPowerProvider, operator)
		if err != nil {
			return errors.Errorf("failed to get operator nonce: %w", err)
		}

		// Build EIP-712 typed data for RegisterOperator
		typedData := apitypes.TypedData{
			Types: apitypes.Types{
				"EIP712Domain": []apitypes.Type{
					{Name: "name", Type: "string"},
					{Name: "version", Type: "string"},
					{Name: "chainId", Type: "uint256"},
					{Name: "verifyingContract", Type: "address"},
				},
				"RegisterOperator": []apitypes.Type{
					{Name: "operator", Type: "address"},
					{Name: "nonce", Type: "uint256"},
				},
			},
			PrimaryType: "RegisterOperator",
			Domain: apitypes.TypedDataDomain{
				Name:              eip712Domain.Name,
				Version:           eip712Domain.Version,
				ChainId:           (*math.HexOrDecimal256)(eip712Domain.ChainId),
				VerifyingContract: eip712Domain.VerifyingContract.Hex(),
			},
			Message: apitypes.TypedDataMessage{
				"operator": operator.Hex(),
				"nonce":    (*math.HexOrDecimal256)(nonce),
			},
		}

		_, preHashedData, err := apitypes.TypedDataAndHash(typedData)
		if err != nil {
			return errors.Errorf("failed to hash typed data: %w", err)
		}

		signature, err := crypto.Sign([]byte(preHashedData), ecdsaPk)
		if err != nil {
			return errors.Errorf("failed to sign: %w", err)
		}

		// Ethereum expects recovery ID to be 27 or 28, not 0 or 1
		if len(signature) == 65 {
			signature[64] += 27
		}

		fmt.Printf("0x%x\n", signature)

		return nil
	},
}
