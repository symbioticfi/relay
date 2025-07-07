package keys

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"log/slog"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	utils_app "middleware-offchain/internal/usecase/utils-app"
)

type config struct {
	path       string
	password   string
	keyTag     uint8
	privateKey string
	generate   bool
	force      bool
}

var cfg config

func NewKeysCmd() (*cobra.Command, error) {
	keysCmd.PersistentFlags().StringVarP(&cfg.path, "path", "p", "./keystore.jks", "path to keystore")
	keysCmd.PersistentFlags().StringVar(&cfg.password, "password", "", "keystore password")

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
}

var printKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

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

			slog.Info("key:", "idx", i, "alias", alias, "public_key", pk.PublicKey())
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

		if cfg.generate {
			cfg.privateKey = "random" // TODO: for each key tag make pk generator
		}

		var err error

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

		key, err := crypto.NewPrivateKey(entity.KeyTag(cfg.keyTag), []byte(cfg.privateKey))
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(entity.KeyTag(cfg.keyTag), key, cfg.password, cfg.force); err != nil {
			return err
		}

		return nil
	},
}

var removeKeyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

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

		if err = keyStore.DeleteKey(entity.KeyTag(cfg.keyTag), cfg.password); err != nil {
			return err
		}

		return nil
	},
}

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

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

		keyTag := entity.KeyTag(cfg.keyTag)
		exists, err := keyStore.HasKey(keyTag)
		if err != nil {
			return err
		}

		if !exists {
			return errors.New("Key doesn't exist")
		}

		key, err := crypto.NewPrivateKey(entity.KeyTag(cfg.keyTag), []byte(cfg.privateKey))
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(entity.KeyTag(cfg.keyTag), key, cfg.password, true); err != nil {
			return err
		}

		return nil
	},
}
