package keys

import (
	"errors"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	cmdhelpers "middleware-offchain/internal/usecase/cmd-helpers"

	"github.com/spf13/cobra"
)

var addKeyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addFlags.PrivateKey == "" && !addFlags.Generate {
			return errors.New("add --generate if private key omitted")
		}

		kt := entity.KeyTag(addFlags.KeyTag)
		return addKey(kt, addFlags.Generate, addFlags.Force, addFlags.PrivateKey)
	},
}

func addKey(keyTag entity.KeyTag, generate bool, force bool, privateKey string) error {
	var err error
	if generate {
		pk, err := crypto.GeneratePrivateKey(keyTag.Type())
		if err != nil {
			return err
		}

		privateKey = string(pk.Bytes())
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

	key, err := crypto.NewPrivateKey(keyTag.Type(), []byte(privateKey))
	if err != nil {
		return err
	}

	if err = keyStore.AddKey(keyTag, key, globalFlags.Password, force); err != nil {
		return err
	}

	return nil
}
