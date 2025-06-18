package keys

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"log/slog"
	"middleware-offchain/core/entity"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/log"
	"syscall"
)

type config struct {
	path       string
	keyTag     uint8
	privateKey string
	generate   bool
	force      bool
	logLevel   string
	logMode    string
}

var cfg config

func NewKeysCmd() (*cobra.Command, error) {
	keysCmd.PersistentFlags().StringVarP(&cfg.path, "path", "p", "./keystore.jks", "path to keystore")
	keysCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "log level")
	keysCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "debug", "log mode")

	addKeyCmd.PersistentFlags().Uint8Var(&cfg.keyTag, "key-tag", 0, "key tag")
	addKeyCmd.PersistentFlags().StringVar(&cfg.privateKey, "private-key", "", "private key")
	addKeyCmd.PersistentFlags().BoolVar(&cfg.generate, "generate", false, "generate key")
	addKeyCmd.PersistentFlags().BoolVar(&cfg.force, "force", false, "force overwrite key")
	if err := addKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		return nil, errors.Errorf("failed to mark key-tag as required: %w", err)
	}

	removeKeyCmd.PersistentFlags().Uint8Var(&cfg.keyTag, "key-tag", 0, "key tag")
	if err := removeKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		return nil, errors.Errorf("failed to mark key-tag as required: %w", err)
	}

	updateKeyCmd.PersistentFlags().Uint8Var(&cfg.keyTag, "key-tag", 0, "key tag")
	updateKeyCmd.PersistentFlags().StringVar(&cfg.privateKey, "private-key", "", "private key")
	if err := updateKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		return nil, errors.Errorf("failed to mark key-tag as required: %w", err)
	}
	if err := updateKeyCmd.MarkPersistentFlagRequired("private-key"); err != nil {
		return nil, errors.Errorf("failed to mark private-key as required: %w", err)
	}

	keysCmd.AddCommand(printKeysCmd)
	keysCmd.AddCommand(addKeyCmd)
	keysCmd.AddCommand(removeKeyCmd)
	keysCmd.AddCommand(updateKeyCmd)

	return keysCmd, nil
}

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Keys tool",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Init(cfg.logLevel, cfg.logMode)
	},
}

var printKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := getPassword()
		if err != nil {
			return err
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, password)
		if err != nil {
			return err
		}

		aliases := keyStore.GetAliases()
		for i, alias := range aliases {
			keyTag, err := keyprovider.AliasToTag(alias)
			if err != nil {
				return err
			}

			pk, err := keyStore.GetPrivateKey(keyTag)
			if err != nil {
				return err
			}

			var publicKeyStr string

			switch keyTag.Type() {
			case entity.KeyTypeBlsBn254:
				keyPair := bls.ComputeKeyPair(pk)
				publicKeyStr = keyPair.PublicKeyG1.String()
			case entity.KeyTypeEcdsaSecp256k1:
				ecdsaPk, err := crypto.ToECDSA(pk)
				if err != nil {
					return err
				}
				publicKeyStr = "E([" + ecdsaPk.PublicKey.X.String() + "," + ecdsaPk.PublicKey.Y.String() + "])"
			default:
				return errors.Errorf("unsupported key tag type: %s", alias)
			}

			slog.Info("key:", "idx", i, "alias", alias, "public_key", publicKeyStr)
		}

		return nil
	},
}

var addKeyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.privateKey == "" && !cfg.generate {
			return errors.New("Add --generate if private key omitted")
		}

		password, err := getPassword()
		if err != nil {
			return err
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, password)
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(entity.KeyTag(cfg.keyTag), common.Hex2Bytes(cfg.privateKey), password, cfg.force); err != nil {
			return err
		}

		return nil
	},
}

var removeKeyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove key",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := getPassword()
		if err != nil {
			return err
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, password)
		if err != nil {
			return err
		}

		if err = keyStore.DeleteKey(entity.KeyTag(cfg.keyTag), password); err != nil {
			return err
		}

		return nil
	},
}

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update key",
	RunE: func(cmd *cobra.Command, args []string) error {
		password, err := getPassword()
		if err != nil {
			return err
		}

		keyStore, err := keyprovider.NewKeystoreProvider(cfg.path, password)
		if err != nil {
			return err
		}

		keyTag := entity.KeyTag(cfg.keyTag)
		exists, err := keyStore.HasKey(keyTag)
		if err != nil {
			return err
		}

		if !exists {
			return errors.New("Key doesn't exist")
		}

		if err = keyStore.AddKey(keyTag, common.Hex2Bytes(cfg.privateKey), password, true); err != nil {
			return err
		}

		return nil
	},
}

func getPassword() (string, error) {
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}
