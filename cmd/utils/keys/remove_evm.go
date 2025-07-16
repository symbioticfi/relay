package keys

import (
	"middleware-offchain/core/entity"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	cmdhelpers "middleware-offchain/internal/usecase/cmd-helpers"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var removeEVMKeyCmd = &cobra.Command{
	Use:   "remove-evm",
	Short: "Remove EVM key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if removeEvmFlags.ChainId == 0 {
			return errors.New("evm chain id omitted")
		}

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

		if err = keyStore.DeleteKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, int(removeEvmFlags.ChainId), globalFlags.Password); err != nil {
			return err
		}

		return nil
	},
}
