package keys

import (
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var removeKeyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if globalFlags.Password == "" {
			globalFlags.Password, err = cmdhelpers.GetPassword()
			if err != nil {
				return err
			}
		}

		keyStore, err := keyprovider.NewKeystoreProvider(globalFlags.Path, globalFlags.Password)
		if err != nil {
			return err
		}

		if removeFlags.EvmNs {
			if removeFlags.ChainID < 0 {
				return errors.New("chain ID is required for evm namespace, use --chain-id=0 for default key for all chains")
			}
			return keyStore.DeleteKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, int(removeFlags.ChainID), globalFlags.Password)
		} else if removeFlags.RelayNs {
			if removeFlags.KeyTag == uint8(symbiotic.KeyTypeInvalid) {
				return errors.New("key tag is required for relay namespace")
			}
			kt := symbiotic.KeyTag(removeFlags.KeyTag)
			if kt.Type() == symbiotic.KeyTypeInvalid {
				return errors.New("invalid key tag, type not supported")
			}
			keyId := kt & 0x0F
			return keyStore.DeleteKeyByNamespaceTypeId(keyprovider.SYMBIOTIC_KEY_NAMESPACE, kt.Type(), int(keyId), globalFlags.Password)
		}
		return errors.New("either --evm-ns or --relay-ns must be specified")
	},
}
