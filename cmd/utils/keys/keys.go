package keys

import (
	"fmt"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	utils_app "middleware-offchain/internal/usecase/utils-app"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func NewKeysCmd() *cobra.Command {
	keysCmd.AddCommand(printKeysCmd)
	keysCmd.AddCommand(addKeyCmd)
	keysCmd.AddCommand(removeKeyCmd)
	keysCmd.AddCommand(updateKeyCmd)

	addFlags()

	return keysCmd
}

var keysCmd = &cobra.Command{
	Use:               "keys",
	Short:             "Keys tool",
	PersistentPreRunE: initConfig,
}

var printKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgFromCtx(cmd.Context())
		var err error

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

		aliases := keyStore.GetAliases()

		fmt.Printf("\nKeys (%d):\n", len(aliases)) // assuming 'keys' is your collection
		fmt.Printf("   # | Alias                | Public Key\n")

		for i, alias := range aliases {
			keyTag, err := keyprovider.AliasToTag(alias)
			if err != nil {
				return err
			}

			pk, err := keyStore.GetPrivateKey(keyTag)
			if err != nil {
				return err
			}

			prettyPk, err := pk.PublicKey().OnChain().MarshalText()
			if err != nil {
				return err
			}

			fmt.Printf("   %d | %-20s | %s\n", i+1, alias, prettyPk)
		}

		return nil
	},
}

var addKeyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgFromCtx(cmd.Context())

		if cfg.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("Key tag omitted")
		}

		if cfg.PrivateKey == "" && !cfg.Generate {
			return errors.New("Add --generate if private key omitted")
		}
		var err error

		kt := entity.KeyTag(cfg.KeyTag)
		if cfg.Generate {
			pk, err := crypto.GeneratePrivateKey(kt)
			if err != nil {
				return err
			}

			cfg.PrivateKey = string(pk.Bytes())
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

		key, err := crypto.NewPrivateKey(kt, []byte(cfg.PrivateKey))
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(kt, key, cfg.Password, cfg.Force); err != nil {
			return err
		}

		return nil
	},
}

var removeKeyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgFromCtx(cmd.Context())
		var err error

		if cfg.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("Key tag omitted")
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

		if err = keyStore.DeleteKey(entity.KeyTag(cfg.KeyTag), cfg.Password); err != nil {
			return err
		}

		return nil
	},
}

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgFromCtx(cmd.Context())
		var err error

		if cfg.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("Key tag omitted")
		}

		if cfg.PrivateKey == "" {
			return errors.New("Private key omitted")
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
		exists, err := keyStore.HasKey(kt)
		if err != nil {
			return err
		}

		if !exists {
			return errors.New("Key doesn't exist")
		}

		key, err := crypto.NewPrivateKey(kt, []byte(cfg.PrivateKey))
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(kt, key, cfg.Password, true); err != nil {
			return err
		}

		return nil
	},
}
