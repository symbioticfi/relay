package operator

import (
	"log/slog"
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

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register operator in core registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())

		// TODO add network opt-in
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

		// duplicate from genesis
		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
		for _, chainId := range evmClient.GetChains() {
			secret, ok := registerFlags.Secrets.Secrets[chainId]
			if !ok {
				secret, _ = privateKeyInput.Show("Enter private key for chain with ID: " + strconv.Itoa(int(chainId)))
			}
			pk, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, common.FromHex(secret))
			if err != nil {
				return err
			}
			err = kp.AddKeyByNamespaceTypeId(
				keyprovider.EVM_KEY_NAMESPACE,
				symbiotic.KeyTypeEcdsaSecp256k1,
				int(chainId),
				pk,
			)
			if err != nil {
				return err
			}
		}

		for _, chainId := range evmClient.GetChains() {
			if _, ok := registerFlags.Secrets.Secrets[chainId]; !ok {
				return errors.Errorf("operator registry in chain %d does not exist", chainId)
			}

			txResult, err := evmClient.RegisterOperator(ctx, operatorRegistries[chainId])
			if err != nil {
				return errors.Errorf("failed to register operator: %w", err)
			}

			slog.InfoContext(ctx, "Operator registered!", "addr", operatorRegistries[chainId], "chain-id", chainId, "txHash", txResult.TxHash.String())
		}

		return nil
	},
}
