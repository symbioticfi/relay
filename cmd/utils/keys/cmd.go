package keys

import (
	"middleware-offchain/core/entity"

	"github.com/spf13/cobra"
)

func NewKeysCmd() *cobra.Command {
	keysCmd.AddCommand(printKeysCmd)
	keysCmd.AddCommand(addKeyCmd)
	keysCmd.AddCommand(addEvmKeyCmd)
	keysCmd.AddCommand(removeKeyCmd)
	keysCmd.AddCommand(updateKeyCmd)

	initFlags()

	return keysCmd
}

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Keys tool",
}

type GlobalFlags struct {
	Path     string
	Password string
}

type AddFlags struct {
	KeyTag     uint8
	PrivateKey string
	Generate   bool
	Force      bool
}

type AddEvmFlags struct {
	ChainId    uint8
	PrivateKey string
	Generate   bool
	Force      bool
}

type RemoveFlags struct {
	KeyTag uint8
}

type RemoveEvmFlags struct {
	ChainId uint8
}

type UpdateFlags struct {
	KeyTag     uint8
	PrivateKey string
}

var globalFlags GlobalFlags
var addFlags AddFlags

var addEvmFlags AddEvmFlags
var removeFlags RemoveFlags
var updateFlags UpdateFlags
var removeEvmFlags RemoveEvmFlags

func initFlags() {
	keysCmd.PersistentFlags().StringVarP(&globalFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	keysCmd.PersistentFlags().StringVar(&globalFlags.Password, "password", "", "Keystore password")

	addKeyCmd.PersistentFlags().Uint8Var(&addFlags.KeyTag, "key-tag", uint8(entity.KeyTypeInvalid), "key tag")
	addKeyCmd.PersistentFlags().StringVar(&addFlags.PrivateKey, "private-key", "", "private key to add in hex")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.Generate, "generate", false, "generate key")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.Force, "force", false, "force overwrite key")
	if err := addKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}

	addEvmKeyCmd.PersistentFlags().Uint8Var(&addEvmFlags.ChainId, "chain-id", 0, "evm chain id")
	addEvmKeyCmd.PersistentFlags().StringVar(&addEvmFlags.PrivateKey, "private-key", "", "private key to add in hex")
	addEvmKeyCmd.PersistentFlags().BoolVar(&addEvmFlags.Generate, "generate", false, "generate random key")
	addEvmKeyCmd.PersistentFlags().BoolVar(&addEvmFlags.Force, "force", false, "force overwrite key")
	if err := addEvmKeyCmd.MarkPersistentFlagRequired("chain-id"); err != nil {
		panic(err)
	}

	removeKeyCmd.PersistentFlags().Uint8Var(&removeFlags.KeyTag, "key-tag", uint8(entity.KeyTypeInvalid), "key tag")
	if err := removeKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}

	removeEVMKeyCmd.PersistentFlags().Uint8Var(&removeEvmFlags.ChainId, "chain-id", 0, "evm chain id")
	if err := removeEVMKeyCmd.MarkPersistentFlagRequired("chain-id"); err != nil {
		panic(err)
	}

	updateKeyCmd.PersistentFlags().Uint8Var(&updateFlags.KeyTag, "key-tag", uint8(entity.KeyTypeInvalid), "key tag")
	updateKeyCmd.PersistentFlags().StringVar(&updateFlags.PrivateKey, "private-key", "", "private key")
	if err := updateKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}
	if err := updateKeyCmd.MarkPersistentFlagRequired("private-key"); err != nil {
		panic(err)
	}
}
