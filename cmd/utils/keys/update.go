package keys

import (
	"github.com/ethereum/go-ethereum/common"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
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

		if updateFlags.PrivateKey == "" {
			return errors.New("private key is required for update")
		}

		privKeyBytes := common.FromHex(updateFlags.PrivateKey)

		if updateFlags.EvmNs {
			if updateFlags.ChainID < 0 {
				return errors.New("chain ID is required for evm namespace, use --chain-id=0 for default key for all chains")
			}
			exists, err := keyStore.HasKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, int(updateFlags.ChainID))
			if err != nil {
				return err
			}

			if !exists {
				return errors.New("key doesn't exist")
			}

			key, err := crypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, privKeyBytes)
			if err != nil {
				return err
			}

			return keyStore.AddKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, int(updateFlags.ChainID), key, globalFlags.Password, true)
		} else if updateFlags.RelayNs {
			if updateFlags.KeyTag == uint8(symbiotic.KeyTypeInvalid) {
				return errors.New("key tag is required for relay namespace")
			}
			kt := symbiotic.KeyTag(updateFlags.KeyTag)
			if kt.Type() == symbiotic.KeyTypeInvalid {
				return errors.New("invalid key tag, type not supported")
			}
			keyId := kt & 0x0F

			exists, err := keyStore.HasKeyByNamespaceTypeId(keyprovider.SYMBIOTIC_KEY_NAMESPACE, kt.Type(), int(keyId))
			if err != nil {
				return err
			}

			if !exists {
				return errors.New("key doesn't exist")
			}

			key, err := crypto.NewPrivateKey(kt.Type(), privKeyBytes)
			if err != nil {
				return err
			}

			return keyStore.AddKeyByNamespaceTypeId(keyprovider.SYMBIOTIC_KEY_NAMESPACE, kt.Type(), int(keyId), key, globalFlags.Password, true)
		} else if updateFlags.P2PNs {
			exists, err := keyStore.HasKeyByNamespaceTypeId(keyprovider.P2P_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, keyprovider.P2P_HOST_IDENTITY_KEY_ID)
			if err != nil {
				return err
			}

			if !exists {
				return errors.New("key doesn't exist")
			}

			key, err := crypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, privKeyBytes)
			if err != nil {
				return err
			}

			return keyStore.AddKeyByNamespaceTypeId(keyprovider.P2P_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, keyprovider.P2P_HOST_IDENTITY_KEY_ID, key, globalFlags.Password, true)
		}
		return errors.New("either --evm or --relay or --p2p must be specified")
	},
}
