package operator

import (
	"strconv"
	"time"

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var registerOperatorCmd = &cobra.Command{
	Use:   "register-operator",
	Short: "Register operator on-chain via VotingPowerProvider",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())

		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		evmClient, err := evm.NewEvmClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: globalFlags.DriverChainId,
				Address: common.HexToAddress(globalFlags.DriverAddress),
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
			Metrics:        metrics.New(metrics.Config{}),
		})
		if err != nil {
			return err
		}

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

		if len(networkConfig.VotingPowerProviders) == 0 {
			return errors.New("no voting power providers found in network config")
		}

		votingPowerProvider := networkConfig.VotingPowerProviders[0]

		// Load the operator key for the voting power provider's chain
		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
		secret, ok := registerOperatorFlags.Secrets.Secrets[votingPowerProvider.ChainId]
		if !ok {
			secret, _ = privateKeyInput.Show("Enter operator private key for chain with ID: " + strconv.Itoa(int(votingPowerProvider.ChainId)))
		}

		pk, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, common.FromHex(secret))
		if err != nil {
			return err
		}
		err = kp.AddKeyByNamespaceTypeId(
			keyprovider.EVM_KEY_NAMESPACE,
			symbiotic.KeyTypeEcdsaSecp256k1,
			int(votingPowerProvider.ChainId),
			pk,
		)
		if err != nil {
			return err
		}

		txResult, err := evmClient.RegisterOperatorVotingPowerProvider(ctx, votingPowerProvider)
		if err != nil {
			return errors.Errorf("failed to register operator: %w", err)
		}

		pterm.Success.Println("Operator registered! TxHash:", txResult.TxHash.String())

		return nil
	},
}
