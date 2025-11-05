package keys

import (
	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
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

		if addFlags.EvmNs {
			if addFlags.ChainID < 0 {
				return errors.New("chain ID is required for evm namespace, use --chain-id=0 for default key for all chains")
			}
			return addKeyWithNamespace(keyprovider.EVM_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, int(addFlags.ChainID), addFlags.Generate, addFlags.Force, addFlags.PrivateKey)
		} else if addFlags.RelayNs {
			if addFlags.KeyTag == uint8(symbiotic.KeyTypeInvalid) {
				return errors.New("key tag is required for relay namespace")
			}
			kt := symbiotic.KeyTag(addFlags.KeyTag)
			if kt.Type() == symbiotic.KeyTypeInvalid {
				return errors.New("invalid key tag, type not supported")
			}
			keyId := kt & 0x0F
			return addKeyWithNamespace(keyprovider.SYMBIOTIC_KEY_NAMESPACE, kt.Type(), int(keyId), addFlags.Generate, addFlags.Force, addFlags.PrivateKey)
		} else if addFlags.P2PNs {
			return addKeyWithNamespace(keyprovider.P2P_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, keyprovider.P2P_HOST_IDENTITY_KEY_ID, addFlags.Generate, addFlags.Force, addFlags.PrivateKey)
		}
		return errors.New("either --evm or --relay or --p2p must be specified")
	},
}

func addKeyWithNamespace(ns string, keyType symbiotic.KeyType, id int, generate bool, force bool, privateKey string) error {
	var err error
	if generate {
		pk, err := crypto.GeneratePrivateKey(keyType)
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

	key, err := crypto.NewPrivateKey(keyType, []byte(privateKey))
	if err != nil {
		return err
	}

	if err = keyStore.AddKeyByNamespaceTypeId(ns, keyType, id, key, globalFlags.Password, force); err != nil {
		return err
	}

	return nil
}
