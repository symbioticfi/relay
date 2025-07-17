package operator

import (
	"fmt"
	"log/slog"
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	symbioticCrypto "middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	"middleware-offchain/internal/usecase/metrics"
	"strconv"
	"time"

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

		client, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: entity.CrossChainAddress{
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
		for _, chainId := range client.GetChains() {
			secret, ok := registerFlags.Secrets.Secrets[chainId]
			if !ok {
				secret, _ = privateKeyInput.Show("Enter private key for chain with ID: " + strconv.Itoa(int(chainId)))
			}
			pk, err := symbioticCrypto.NewPrivateKey(entity.KeyTypeEcdsaSecp256k1, common.FromHex(secret))
			if err != nil {
				return err
			}
			err = kp.AddKeyByNamespaceTypeId(
				keyprovider.EVM_KEY_NAMESPACE,
				entity.KeyTypeEcdsaSecp256k1,
				int(chainId),
				pk,
			)
			if err != nil {
				return err
			}
		}

		for _, chainId := range client.GetChains() {
			if _, ok := registerFlags.Secrets.Secrets[chainId]; !ok {
				return fmt.Errorf("operator registry in chain %d does not exist", chainId)
			}

			txResult, err := client.RegisterOperator(ctx, operatorRegistries[chainId])
			if err != nil {
				return errors.Errorf("failed to register operator: %w", err)
			}

			slog.InfoContext(ctx, "Operator registered!", "addr", operatorRegistries[chainId], "chain-id", chainId, "txHash", txResult.TxHash.String())
		}

		return nil
	},
}
