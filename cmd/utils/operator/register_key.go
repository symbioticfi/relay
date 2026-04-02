package operator

import (
	"context"
	"log/slog"
	"slices"
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

type registerKeyBuilder interface {
	Register(
		ctx context.Context,
		pk symbioticCrypto.PrivateKey,
		kt symbiotic.KeyTag,
		operatorAddress common.Address,
	) (symbiotic.TxResult, error)
}

var newRegisterKeyMetrics = func() *metrics.Metrics {
	return metrics.New(metrics.Config{})
}

var newRegisterKeyEVMClient = func(ctx context.Context, cfg evm.Config) (evm.IEvmClient, error) {
	return evm.NewEvmClient(ctx, cfg)
}

var newRegisterKeyBuilder = func(evmClient evm.IEvmClient) (registerKeyBuilder, error) {
	return key_registerer.NewRegisterer(key_registerer.Config{EVMClient: evmClient})
}

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

		evmClient, err := newRegisterKeyEVMClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: globalFlags.DriverChainId,
				Address: common.HexToAddress(globalFlags.DriverAddress),
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
			Metrics:        newRegisterKeyMetrics(),
		})
		if err != nil {
			return err
		}

		chainId, err := registerKeyChainID(ctx, evmClient)
		if err != nil {
			return errors.Errorf("failed to resolve key registry chain: %w", err)
		}
		if !hasConfiguredChain(evmClient.GetChains(), chainId) {
			return errors.Errorf("keys provider chain %d is not configured", chainId)
		}

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

		keyReg, err := newRegisterKeyBuilder(evmClient)
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

func registerKeyChainID(ctx context.Context, evmClient evm.IEvmClient) (uint64, error) {
	currentOnchainEpoch, err := evmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return 0, errors.Errorf("failed to get current epoch: %w", err)
	}

	captureTimestamp, err := evmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return 0, errors.Errorf("failed to get capture timestamp: %w", err)
	}

	networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
	if err != nil {
		return 0, errors.Errorf("failed to get config: %w", err)
	}

	return networkConfig.KeysProvider.ChainId, nil
}

func hasConfiguredChain(configuredChainIDs []uint64, targetChainID uint64) bool {
	return slices.Contains(configuredChainIDs, targetChainID)
}
