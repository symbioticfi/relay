package operator

import (
	"context"
	"fmt"
	"log/slog"
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	utils_app "middleware-offchain/internal/usecase/utils-app"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var operatorRegistries = map[uint64]entity.CrossChainAddress{
	111: {
		ChainId: 111,
		Address: common.HexToAddress("0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"),
	},
}

func NewOperatorCmd() *cobra.Command {
	operatorCmd.AddCommand(infoCmd)
	operatorCmd.AddCommand(registerCmd)
	operatorCmd.AddCommand(registerKeyCmd)

	addFlags()

	return operatorCmd
}

var operatorCmd = &cobra.Command{
	Use:               "operator",
	Short:             "Operator tool",
	PersistentPreRunE: initConfig,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print operator information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(cmd.Context())
		cfg := cfgFromCtx(ctx)

		if cfg.Path == "" {
			return errors.New("Keystore path is required")
		}

		if cfg.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("Key tag omitted")
		}

		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		client, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs: cfg.ChainsUrl,
			DriverAddress: entity.CrossChainAddress{
				Address: common.HexToAddress(cfg.Driver.Address),
				ChainId: cfg.Driver.ChainID,
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
		})
		if err != nil {
			return errors.Errorf("Failed to init evm client: %w", err)
		}

		if cfg.Password == "" {
			cfg.Password, err = utils_app.GetPassword()
			if err != nil {
				return errors.Errorf("Failed to get password: %w", err)
			}
		}

		if cfg.Epoch == 0 {
			cfg.Epoch, err = client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("Failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := client.GetEpochStart(ctx, cfg.Epoch)
		if err != nil {
			return errors.Errorf("Failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
		}

		epoch, err := client.GetLastCommittedHeaderEpoch(ctx, networkConfig.Replicas[0])
		if err != nil {
			return errors.Errorf("Failed to get valset header: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("Failed to create valset deriver: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set: %w", err)
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.Path, cfg.Password)
		if err != nil {
			return err
		}

		kt := entity.KeyTag(cfg.KeyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		validator, found := valset.FindValidatorByKey(kt, pk.PublicKey().Raw())
		if !found {
			return errors.Errorf("Validator not found for key: %d %s", kt, common.Bytes2Hex(pk.PublicKey().Raw()))
		}

		fmt.Print("Operator Info")

		str, err := utils_app.MarshalTextValidator(validator, cfg.Compact)
		if err != nil {
			return errors.Errorf("Failed to log validator: %w", err)
		}

		fmt.Print(str)

		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register operator in core registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())
		cfg := cfgFromCtx(ctx)

		if len(cfg.ChainsId) != 1 {
			return errors.New("Only single chain is supported")
		}

		if cfg.PrivateKey == "" {
			return errors.New("Private key is required")
		}

		if cfg.Path == "" {
			return errors.New("Keystore path is required")
		}

		if cfg.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("Key tag omitted")
		}
		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		client, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs: cfg.ChainsUrl,
			DriverAddress: entity.CrossChainAddress{
				Address: common.HexToAddress(cfg.Driver.Address),
				ChainId: cfg.Driver.ChainID,
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
		})
		if err != nil {
			return errors.Errorf("Failed to init evm client: %w", err)
		}

		txResult, err := client.RegisterOperator(ctx, operatorRegistries[cfg.ChainsId[0]])
		if err != nil {
			return errors.Errorf("Failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "Operator registered!", "addr", operatorRegistries[cfg.ChainsId[0]], "txHash", txResult.TxHash.String())

		return nil
	},
}

var registerKeyCmd = &cobra.Command{
	Use:   "register-key",
	Short: "Register operator key in key registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(cmd.Context())
		cfg := cfgFromCtx(ctx)

		if len(cfg.ChainsId) != 1 {
			return errors.New("Only single chain is supported")
		}

		if cfg.PrivateKey == "" {
			return errors.New("Private key is required")
		}
		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		client, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs: cfg.ChainsUrl,
			DriverAddress: entity.CrossChainAddress{
				Address: common.HexToAddress(cfg.Driver.Address),
				ChainId: cfg.Driver.ChainID,
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
		})
		if err != nil {
			return errors.Errorf("Failed to init evm client: %w", err)
		}

		if cfg.Password == "" {
			cfg.Password, err = utils_app.GetPassword()
			if err != nil {
				return err
			}
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.Path, cfg.Password)
		if err != nil {
			return err
		}

		kt := entity.KeyTag(cfg.KeyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		currentOnchainEpoch, err := client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("Failed to get current epoch: %w", err)
		}

		captureTimestamp, err := client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("Failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
		}

		eip712Domain, err := client.GetEip712Domain(ctx, networkConfig.KeysProvider)
		if err != nil {
			return errors.Errorf("Failed to get eip712 domain: %w", err)
		}

		ecdsaPk, err := crypto.HexToECDSA(cfg.PrivateKey)
		if err != nil {
			return errors.Errorf("Failed to parse private key: %w", err)
		}

		operator := crypto.PubkeyToAddress(ecdsaPk.PublicKey)
		key := pk.PublicKey().OnChain()
		commitmentData, err := keyCommitmentData(eip712Domain, operator, crypto.Keccak256(key))
		if err != nil {
			return errors.Errorf("Failed to get commitment data: %w", err)
		}

		signature, _, err := pk.Sign(commitmentData)
		if err != nil {
			return errors.Errorf("Failed to sign commitment data: %w", err)
		}

		var extraData []byte
		if kt.Type() == entity.KeyTypeBlsBn254 {
			extraData = pk.PublicKey().Raw()[32:]
		}

		txResult, err := client.RegisterKey(ctx, networkConfig.KeysProvider, kt, key, signature, extraData)
		if err != nil {
			return errors.Errorf("Failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "Operator registered!", "addr", operatorRegistries[cfg.ChainsId[0]], "txHash", txResult.TxHash.String())

		return nil
	},
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("Received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
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
