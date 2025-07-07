package operator

import (
	"context"
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

type config struct {
	epoch         uint64
	rpcURL        string
	driverAddress string
	path          string
	password      string
	keyTag        uint8
	privateKey    string
	compact       bool
	chainId       uint64

	driverCrossChainAddress entity.CrossChainAddress
	client                  *evm.Client
}

var operatorRegistries = map[uint64]entity.CrossChainAddress{
	111: {
		ChainId: 111,
		Address: common.HexToAddress("0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"),
	},
}

var cfg config

func NewOperatorCmd() (*cobra.Command, error) {
	operatorCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	operatorCmd.PersistentFlags().StringVar(&cfg.driverAddress, "driver-address", "", "Driver contract address")
	if err := operatorCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return nil, errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := operatorCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		return nil, errors.Errorf("failed to mark driver-address as required: %w", err)
	}
	infoCmd.PersistentFlags().Uint64Var(&cfg.epoch, "epoch", 0, "Network epoch")
	infoCmd.PersistentFlags().StringVar(&cfg.path, "path", "", "Keystore path")
	if err := infoCmd.MarkPersistentFlagRequired("path"); err != nil {
		return nil, errors.Errorf("failed to mark path as required: %w", err)
	}
	infoCmd.PersistentFlags().StringVar(&cfg.password, "password", "", "Keystore password")
	infoCmd.PersistentFlags().Uint8Var(&cfg.keyTag, "key-tag", 0, "Key tag of operator key")
	if err := infoCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		return nil, errors.Errorf("failed to mark key-tag as required: %w", err)
	}
	infoCmd.PersistentFlags().BoolVar(&cfg.compact, "compact", false, "Compact operator info print")

	registerCmd.PersistentFlags().StringVar(&cfg.privateKey, "private-key", "", "Private key of operator")
	if err := registerKeyCmd.MarkPersistentFlagRequired("private-key"); err != nil {
		return nil, errors.Errorf("failed to mark private-key as required: %w", err)
	}
	registerCmd.PersistentFlags().Uint64Var(&cfg.chainId, "chain-id", 0, "Chain id where to register")
	if err := registerKeyCmd.MarkPersistentFlagRequired("chain-id"); err != nil {
		return nil, errors.Errorf("failed to mark chain-id as required: %w", err)
	}

	registerKeyCmd.PersistentFlags().StringVar(&cfg.privateKey, "private-key", "", "Private key of operator")
	if err := registerKeyCmd.MarkPersistentFlagRequired("private-key"); err != nil {
		return nil, errors.Errorf("failed to mark private-key as required: %w", err)
	}
	registerKeyCmd.PersistentFlags().StringVar(&cfg.path, "path", "", "Keystore path")
	if err := registerKeyCmd.MarkPersistentFlagRequired("path"); err != nil {
		return nil, errors.Errorf("failed to mark path as required: %w", err)
	}
	registerKeyCmd.PersistentFlags().StringVar(&cfg.password, "password", "", "Keystore password")
	registerKeyCmd.PersistentFlags().Uint8Var(&cfg.keyTag, "key-tag", 0, "Key tag of operator key")
	if err := registerKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		return nil, errors.Errorf("failed to mark key-tag as required: %w", err)
	}

	operatorCmd.AddCommand(infoCmd)
	operatorCmd.AddCommand(registerCmd)
	operatorCmd.AddCommand(registerKeyCmd)

	return operatorCmd, nil
}

var operatorCmd = &cobra.Command{
	Use:   "network",
	Short: "Network tool",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(context.Background())

		var err error
		var privateKey []byte

		if cfg.privateKey != "" {
			privateKey = common.Hex2Bytes(cfg.privateKey)
		}

		cfg.driverCrossChainAddress = entity.CrossChainAddress{ChainId: 111, Address: common.HexToAddress(cfg.driverAddress)}
		cfg.client, err = evm.NewEVMClient(ctx, evm.Config{
			Chains: []entity.ChainURL{{
				ChainID: 111,
				RPCURL:  cfg.rpcURL,
			}},
			DriverAddress:  cfg.driverCrossChainAddress,
			RequestTimeout: time.Second * 5,
			PrivateKey:     privateKey,
		})

		if err != nil {
			return errors.Errorf("failed to create symbiotic client: %w", err)
		}

		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print operator information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(context.Background())

		if cfg.password == "" {
			cfg.password, err = utils_app.GetPassword()
			if err != nil {
				return err
			}
		}

		if cfg.epoch == 0 {
			cfg.epoch, err = cfg.client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := cfg.client.GetEpochStart(ctx, cfg.epoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := cfg.client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		epoch, err := cfg.client.GetLastCommittedHeaderEpoch(ctx, networkConfig.Replicas[0])
		if err != nil {
			return errors.Errorf("failed to get valset header: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(cfg.client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set: %w", err)
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, cfg.password)
		if err != nil {
			return err
		}

		kt := entity.KeyTag(cfg.keyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		validator, found := valset.FindValidatorByKey(kt, pk.PublicKey().Raw())
		if !found {
			return errors.Errorf("validator not found for key: %d %s", kt, common.Bytes2Hex(pk.PublicKey().Raw()))
		}

		slog.InfoContext(ctx, "Operator Info")

		if err := utils_app.LogValidator(ctx, validator, cfg.compact); err != nil {
			return errors.Errorf("failed to log validator: %w", err)
		}

		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register operator in core registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(context.Background())

		txResult, err := cfg.client.RegisterOperator(ctx, operatorRegistries[cfg.chainId])
		if err != nil {
			return errors.Errorf("failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "operator registered!", "addr", operatorRegistries[cfg.chainId], "txHash", txResult.TxHash.String())

		return nil
	},
}

var registerKeyCmd = &cobra.Command{
	Use:   "register-key",
	Short: "Register operator key in key registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(context.Background())

		if cfg.password == "" {
			cfg.password, err = utils_app.GetPassword()
			if err != nil {
				return err
			}
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, cfg.password)
		if err != nil {
			return err
		}

		kt := entity.KeyTag(cfg.keyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		currentOnchainEpoch, err := cfg.client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		captureTimestamp, err := cfg.client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := cfg.client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		eip712Domain, err := cfg.client.GetEip712Domain(ctx, networkConfig.KeysProvider)
		if err != nil {
			return errors.Errorf("failed to get eip712 domain: %w", err)
		}

		ecdsaPk, err := crypto.HexToECDSA(cfg.privateKey)
		if err != nil {
			return errors.Errorf("failed to parse private key: %w", err)
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

		txResult, err := cfg.client.RegisterKey(ctx, networkConfig.KeysProvider, kt, key, signature, extraData)
		if err != nil {
			return errors.Errorf("failed to register operator: %w", err)
		}

		slog.InfoContext(ctx, "operator registered!", "addr", operatorRegistries[cfg.chainId], "txHash", txResult.TxHash.String())

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
		slog.Info("received signal", "signal", sig)
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
