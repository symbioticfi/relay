package keys

import (
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/spf13/cobra"
)

func NewKeysCmd() *cobra.Command {
	keysCmd.AddCommand(printKeysCmd)
	keysCmd.AddCommand(addKeyCmd)
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
	EvmNs      bool
	RelayNs    bool
	P2PNs      bool
	KeyTag     uint8
	ChainID    int64
	PrivateKey string
	Generate   bool
	Force      bool
}

type RemoveFlags struct {
	EvmNs   bool
	RelayNs bool
	P2PNs   bool
	KeyTag  uint8
	ChainID int64
}

type UpdateFlags struct {
	EvmNs      bool
	RelayNs    bool
	P2PNs      bool
	KeyTag     uint8
	ChainID    int64
	PrivateKey string
	Force      bool
}

var globalFlags GlobalFlags
var addFlags AddFlags

var removeFlags RemoveFlags
var updateFlags UpdateFlags

func initFlags() {
	keysCmd.PersistentFlags().StringVarP(&globalFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	keysCmd.PersistentFlags().StringVar(&globalFlags.Password, "password", "", "Keystore password")

	addKeyCmd.PersistentFlags().BoolVar(&addFlags.EvmNs, "evm", false, "use evm namespace keys")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.RelayNs, "relay", false, "use relay namespace keys")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.P2PNs, "p2p", false, "use p2p key")
	addKeyCmd.PersistentFlags().Uint8Var(&addFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag for relay keys")
	addKeyCmd.PersistentFlags().Int64Var(&addFlags.ChainID, "chain-id", -1, "chain id for evm keys, use 0 for default key for all chains")
	addKeyCmd.PersistentFlags().StringVar(&addFlags.PrivateKey, "private-key", "", "private key to add in hex")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.Generate, "generate", false, "generate key")
	addKeyCmd.PersistentFlags().BoolVar(&addFlags.Force, "force", false, "force overwrite key")

	removeKeyCmd.PersistentFlags().BoolVar(&removeFlags.EvmNs, "evm", false, "use evm namespace keys")
	removeKeyCmd.PersistentFlags().BoolVar(&removeFlags.RelayNs, "relay", false, "use relay namespace keys")
	removeKeyCmd.PersistentFlags().Uint8Var(&removeFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag for relay keys")
	removeKeyCmd.PersistentFlags().Int64Var(&removeFlags.ChainID, "chain-id", -1, "chain id for evm keys, use 0 for default key for all chains")
	removeKeyCmd.PersistentFlags().BoolVar(&removeFlags.P2PNs, "p2p", false, "use p2p key")

	updateKeyCmd.PersistentFlags().BoolVar(&updateFlags.EvmNs, "evm", false, "use evm namespace keys")
	updateKeyCmd.PersistentFlags().BoolVar(&updateFlags.RelayNs, "relay", false, "use relay namespace keys")
	updateKeyCmd.PersistentFlags().Uint8Var(&updateFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag for relay keys")
	updateKeyCmd.PersistentFlags().Int64Var(&updateFlags.ChainID, "chain-id", -1, "chain id for evm keys, use 0 for default key for all chains")
	updateKeyCmd.PersistentFlags().StringVar(&updateFlags.PrivateKey, "private-key", "", "private key to add in hex")
	updateKeyCmd.PersistentFlags().BoolVar(&updateFlags.Force, "force", false, "force overwrite key")
	updateKeyCmd.PersistentFlags().BoolVar(&updateFlags.P2PNs, "p2p", false, "use p2p key")
}
