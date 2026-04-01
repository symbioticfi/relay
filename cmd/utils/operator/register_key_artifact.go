package operator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	key_registerer "github.com/symbioticfi/relay/internal/usecase/key-registerer"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type registerKeyArtifactBuilder interface {
	BuildArtifact(ctx context.Context, pk symbioticCrypto.PrivateKey, kt symbiotic.KeyTag, operatorAddress common.Address) (key_registerer.RegistrationArtifact, error)
	Register(ctx context.Context, pk symbioticCrypto.PrivateKey, kt symbiotic.KeyTag, operatorAddress common.Address) (symbiotic.TxResult, error)
}

type registerKeyBuildConfig struct {
	EVMClient evm.IEvmClient
}

var newRegisterKeyArtifactMetrics = func() *metrics.Metrics {
	return metrics.New(metrics.Config{})
}

var newRegisterKeyArtifactBuilder = func(cfg registerKeyBuildConfig) (registerKeyArtifactBuilder, error) {
	return key_registerer.NewRegisterer(key_registerer.Config{EVMClient: cfg.EVMClient})
}

var registerKeyArtifactCmd = &cobra.Command{
	Use:   "register-key-artifact",
	Short: "Build operator key registration artifact without submitting a transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())

		if registerKeyArtifactFlags.OperatorAddress == "" {
			return errors.New("--operator-address is required")
		}
		if !common.IsHexAddress(registerKeyArtifactFlags.OperatorAddress) {
			return errors.New("--operator-address must be a valid hex address")
		}

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
			Metrics:        newRegisterKeyArtifactMetrics(),
		})
		if err != nil {
			return err
		}

		kt := symbiotic.KeyTag(registerKeyArtifactFlags.KeyTag)
		pk, err := loadRegisterKeyArtifactPrivateKey(registerKeyArtifactFlags.Path, &registerKeyArtifactFlags.Password, kt)
		if err != nil {
			return err
		}

		keyReg, err := newRegisterKeyArtifactBuilder(registerKeyBuildConfig{
			EVMClient: evmClient,
		})
		if err != nil {
			return errors.Errorf("failed to create registerer: %w", err)
		}

		artifact, err := keyReg.BuildArtifact(ctx, pk, kt, common.HexToAddress(registerKeyArtifactFlags.OperatorAddress))
		if err != nil {
			return errors.Errorf("failed to build registration artifact: %w", err)
		}

		encodedArtifact, err := json.Marshal(artifact)
		if err != nil {
			return errors.Errorf("failed to encode registration artifact: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(append(encodedArtifact, '\n'))
		return err
	},
}

func loadRegisterKeyArtifactPrivateKey(path string, password *string, kt symbiotic.KeyTag) (symbioticCrypto.PrivateKey, error) {
	if *password == "" {
		pass, err := cmdhelpers.GetPassword()
		if err != nil {
			return nil, err
		}
		*password = pass
	}

	keyStore, err := keyprovider.NewKeystoreProvider(path, *password)
	if err != nil {
		return nil, err
	}

	pk, err := keyStore.GetPrivateKey(kt)
	if err != nil {
		return nil, errors.Errorf("failed to get private key  for keyTag %v from keystore: %w", kt, err)
	}
	return pk, nil
}
