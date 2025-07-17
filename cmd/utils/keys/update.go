package keys

import (
	"errors"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	cmdhelpers "middleware-offchain/internal/usecase/cmd-helpers"

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

		if err = keyStore.AddKey(kt, key, globalFlags.Password, true); err != nil {
			return err
		}

		return nil
	},
}
