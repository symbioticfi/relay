package operator

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	key_registerer "github.com/symbioticfi/relay/internal/usecase/key-registerer"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

var registerKeyCmd = &cobra.Command{
	Use:   "register-key",
	Short: "Register operator key in key registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
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

		// TODO multiple chains key registration support
		if len(evmClient.GetChains()) != 1 {
			return errors.New("only single chain is supported")
		}
		chainId := evmClient.GetChains()[0]

		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
		secret, ok := registerKeyFlags.Secrets.Secrets[chainId]
		if !ok {
			secret, _ = privateKeyInput.Show("Enter private key for chain with ID: " + strconv.Itoa(int(chainId)))
		}
		evmPK, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, common.FromHex(secret))
		if err != nil {
			return err
		}
		err = kp.AddKeyByNamespaceTypeId(
			keyprovider.EVM_KEY_NAMESPACE,
			symbiotic.KeyTypeEcdsaSecp256k1,
			int(chainId),
			evmPK,
		)
		if err != nil {
			return err
		}

		if registerKeyFlags.Password == "" {
			registerKeyFlags.Password, err = cmdhelpers.GetPassword()
			if err != nil {
				return err
			}
		}

		keyStore, err := keyprovider.NewKeystoreProvider(registerKeyFlags.Path, registerKeyFlags.Password)
		if err != nil {
			return err
		}

		kt := symbiotic.KeyTag(registerKeyFlags.KeyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return errors.Errorf("failed to get private key  for keyTag %v from keystore: %w", kt, err)
		}

		ecdsaPk, err := crypto.HexToECDSA(secret)
		if err != nil {
			return err
		}
		operator := crypto.PubkeyToAddress(ecdsaPk.PublicKey)

		keyReg, err := key_registerer.NewRegisterer(key_registerer.Config{
			EVMClient: evmClient,
		})
		if err != nil {
			return errors.Errorf("failed to create registerer: %w", err)
		}

		// Use the adjusted signature for registration
		txResult, err := keyReg.Register(ctx, pk, kt, operator)
		if err != nil {
			return errors.Errorf("failed to register key: %w", err)
		}

		slog.InfoContext(ctx, "Operator Key registered!", "txHash", txResult.TxHash.String(), "key-tag", kt)

		return nil
	},
}
