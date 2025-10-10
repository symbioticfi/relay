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

		if removeFlags.KeyTag == uint8(symbiotic.KeyTypeInvalid) {
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

		if err = keyStore.DeleteKey(symbiotic.KeyTag(removeFlags.KeyTag), globalFlags.Password); err != nil {
			return err
		}

		return nil
	},
}
