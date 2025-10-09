package keys

import (
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update key",
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

		kt := entity.KeyTag(updateFlags.KeyTag)
		exists, err := keyStore.HasKey(kt)
		if err != nil {
			return err
		}

		if !exists {
			return errors.New("key doesn't exist")
		}

		key, err := crypto.NewPrivateKey(kt.Type(), []byte(updateFlags.PrivateKey))
		if err != nil {
			return err
		}

		if err = keyStore.AddKey(updateFlags.Namespace, kt, key, globalFlags.Password, true); err != nil {
			return err
		}

		return nil
	},
}
