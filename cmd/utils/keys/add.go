package keys

import (
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
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
		return addKey(addFlags.Namespace, kt, addFlags.Generate, addFlags.Force, addFlags.PrivateKey)
	},
}

func addKey(namespace string, keyTag entity.KeyTag, generate bool, force bool, privateKey string) error {
	var err error
	if generate {
		pk, err := crypto.GeneratePrivateKey(keyTag.Type())
		if err != nil {
			return err
		}

		privateKey = string(pk.Bytes())
	} else {
		pkBytes := common.FromHex(privateKey)
		if len(pkBytes) == 0 {
			return errors.New("private key cannot be empty")
		}
		privateKey = string(pkBytes)
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

	if err = keyStore.AddKey(namespace, keyTag, key, globalFlags.Password, force); err != nil {
		return err
	}

	return nil
}
