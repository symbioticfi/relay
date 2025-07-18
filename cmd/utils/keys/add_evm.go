package keys

import (
	"errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var addEvmKeyCmd = &cobra.Command{
	Use:   "add-evm",
	Short: "Add EVM key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addEvmFlags.PrivateKey == "" && !addEvmFlags.Generate {
			return errors.New("add --generate if private key omitted")
		}

		if addEvmFlags.ChainId == 0 && !addEvmFlags.DefaultKey {
			return errors.New("evm chain id omitted, either pass chain-id or use --default-key flag")
		}

		if addEvmFlags.DefaultKey {
			addEvmFlags.ChainId = keyprovider.DEFAULT_EVM_CHAIN_ID
		}

		return addKeyWithNamespace(keyprovider.EVM_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, int(addEvmFlags.ChainId), addEvmFlags.Generate, addEvmFlags.Force, addEvmFlags.PrivateKey)
	},
}

func addKeyWithNamespace(ns string, keyTag entity.KeyType, id int, generate bool, force bool, privateKey string) error {
	var err error
	if generate {
		pk, err := crypto.GeneratePrivateKey(keyTag)
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

	key, err := crypto.NewPrivateKey(keyTag, []byte(privateKey))
	if err != nil {
		return err
	}

	if err = keyStore.AddKeyByNamespaceTypeId(ns, keyTag, id, key, globalFlags.Password, force); err != nil {
		return err
	}

	return nil
}
