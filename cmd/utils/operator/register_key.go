package operator

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/math"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/bls12381"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"

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

		// duplicate from genesis
		privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
		secret, ok := registerKeyFlags.Secrets.Secrets[chainId]
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
		pk, err = keyStore.GetPrivateKey(kt)
		if err != nil {
			return errors.Errorf("failed to get private key  for keyTag %v from keystore: %w", kt, err)
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

		eip712Domain, err := evmClient.GetEip712Domain(ctx, networkConfig.KeysProvider)
		if err != nil {
			return errors.Errorf("failed to get eip712 domain: %w", err)
		}

		ecdsaPk, err := crypto.HexToECDSA(secret)
		if err != nil {
			return err
		}
		operator := crypto.PubkeyToAddress(ecdsaPk.PublicKey)

		key := pk.PublicKey().OnChain()

		commitmentData, err := keyCommitmentData(eip712Domain, operator, key)
		if err != nil {
			return errors.Errorf("failed to get commitment data: %w", err)
		}

		signature, _, err := pk.Sign(commitmentData)
		if err != nil {
			return errors.Errorf("failed to sign commitment data: %w", err)
		}

		// For ECDSA signatures, we need to adjust the recovery ID
		// Go's crypto.Sign() returns V as 0 or 1, but Ethereum expects 27 or 28
		if kt.Type() == symbiotic.KeyTypeEcdsaSecp256k1 && len(signature) == 65 {
			// Convert recovery ID from 0/1 to 27/28 for Ethereum
			signature[64] += 27
		}

		var extraData []byte
		switch kt.Type() {
		case symbiotic.KeyTypeBlsBn254:
			blsKey, err := blsBn254.FromRaw(pk.PublicKey().Raw())
			if err != nil {
				return errors.Errorf("failed to parse BLS public key: %w", err)
			}
			rawByte := blsKey.G2().RawBytes()
			extraData = rawByte[:]
		case symbiotic.KeyTypeBls12381:
			blsKey, err := bls12381.FromRaw(pk.PublicKey().Raw())
			if err != nil {
				return errors.Errorf("failed to parse BLS public key: %w", err)
			}
			rawByte := blsKey.G2().RawBytes()
			extraData = rawByte[:]
		case symbiotic.KeyTypeEcdsaSecp256k1:
			// no extra data needed for ECDSA keys
		case symbiotic.KeyTypeInvalid:
			return errors.New("invalid key type")
		}

		// Use the adjusted signature for registration
		txResult, err := evmClient.RegisterKey(ctx, networkConfig.KeysProvider, kt, key, signature, extraData)
		if err != nil {
			return errors.Errorf("failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "Operator Key registered!", "txHash", txResult.TxHash.String(), "key-tag", kt)

		return nil
	},
}

func keyCommitmentData(eip712Domain symbiotic.Eip712Domain, operator common.Address, keyBytes []byte) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"KeyOwnership": []apitypes.Type{
				{Name: "operator", Type: "address"},
				{Name: "key", Type: "bytes"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:              eip712Domain.Name,
			Version:           eip712Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(eip712Domain.ChainId),
			VerifyingContract: eip712Domain.VerifyingContract.Hex(),
		},
		PrimaryType: "KeyOwnership",
		Message: map[string]interface{}{
			"operator": operator.Hex(),
			"key":      keyBytes,
		},
	}

	_, preHashedData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	return []byte(preHashedData), nil
}
