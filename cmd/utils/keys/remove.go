package keys

import (
	"github.com/symbioticfi/relay/core/entity"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"

	"github.com/spf13/cobra"

	"errors"
)

var removeKeyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if removeFlags.KeyTag == uint8(entity.KeyTypeInvalid) {
			return errors.New("key tag omitted")
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

		if err = keyStore.DeleteKey(entity.KeyTag(removeFlags.KeyTag), globalFlags.Password); err != nil {
			return err
		}

		return nil
	},
}
