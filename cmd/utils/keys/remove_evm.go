package keys

import (
	"github.com/symbioticfi/relay/core/entity"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var removeEVMKeyCmd = &cobra.Command{
	Use:   "remove-evm",
	Short: "Remove EVM key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if removeEvmFlags.ChainId == 0 && !removeEvmFlags.DefaultKey {
			return errors.New("evm chain id omitted, either pass chain-id or use --default-key flag")
		}

		if globalFlags.Password == "" {
			globalFlags.Password, err = cmdhelpers.GetPassword()
			if err != nil {
				return err
			}
		}

		if removeEvmFlags.DefaultKey {
			removeEvmFlags.ChainId = keyprovider.DEFAULT_EVM_CHAIN_ID
		}

		keyStore, err := keyprovider.NewKeystoreProvider(globalFlags.Path, globalFlags.Password)
		if err != nil {
			return err
		}

		if err = keyStore.DeleteKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, int(removeEvmFlags.ChainId), globalFlags.Password); err != nil {
			return err
		}

		return nil
	},
}
