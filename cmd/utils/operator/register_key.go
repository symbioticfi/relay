package operator

import (
	"log/slog"
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	symbioticCrypto "middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	cmdhelpers "middleware-offchain/internal/usecase/cmd-helpers"
	"middleware-offchain/internal/usecase/metrics"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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

		// TODO multiple chains key registration support
		if len(client.GetChains()) != 1 {
			return errors.New("only single chain is supported")
		}
		chainId := client.GetChains()[0]

		// duplicate from genesis
		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
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

		kt := entity.KeyTag(registerKeyFlags.KeyTag)
		pk, err = keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		currentOnchainEpoch, err := client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		captureTimestamp, err := client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		eip712Domain, err := client.GetEip712Domain(ctx, networkConfig.KeysProvider)
		if err != nil {
			return errors.Errorf("failed to get eip712 domain: %w", err)
		}

		ecdsaPk, err := crypto.HexToECDSA(secret)
		if err != nil {
			return err
		}
		operator := crypto.PubkeyToAddress(ecdsaPk.PublicKey)
		key := pk.PublicKey().OnChain()
		commitmentData, err := keyCommitmentData(eip712Domain, operator, crypto.Keccak256(key))
		if err != nil {
			return errors.Errorf("failed to get commitment data: %w", err)
		}

		signature, _, err := pk.Sign(commitmentData)
		if err != nil {
			return errors.Errorf("failed to sign commitment data: %w", err)
		}

		var extraData []byte
		if kt.Type() == entity.KeyTypeBlsBn254 {
			extraData = pk.PublicKey().Raw()[32:]
		}

		txResult, err := client.RegisterKey(ctx, networkConfig.KeysProvider, kt, key, signature, extraData)
		if err != nil {
			return errors.Errorf("failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "Operator registered!", "addr", operatorRegistries[chainId], "txHash", txResult.TxHash.String())

		return nil
	},
}

func keyCommitmentData(eip712Domain entity.Eip712Domain, operator common.Address, keyHash []byte) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
			},
			"KeyOwnership": []apitypes.Type{
				{Name: "operator", Type: "address"},
				{Name: "key", Type: "bytes"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:    eip712Domain.Name,
			Version: eip712Domain.Version,
		},
		PrimaryType: "KeyOwnership",
		Message: map[string]interface{}{
			"operator": operator,
			"key":      keyHash,
		},
	}

	_, data, err := apitypes.TypedDataAndHash(typedData)
	return []byte(data), err
}
