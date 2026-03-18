package keys

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

var signKeyCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign a message with a relay key",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if !cmd.Flags().Changed("key-tag") {
			return errors.New("key tag is required")
		}

		keyTag := symbiotic.KeyTag(signFlags.KeyTag)
		if keyTag.Type() == symbiotic.KeyTypeInvalid {
			return errors.New("invalid key tag, type not supported")
		}

		if signFlags.MessageHex == "" {
			return errors.New("message hex is required")
		}

		messageBytes, err := hexutil.Decode(signFlags.MessageHex)
		if err != nil {
			return errors.Errorf("invalid message hex: %w", err)
		}
		if len(messageBytes) == 0 {
			return errors.New("message hex cannot be empty")
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

		privateKey, err := keyStore.GetPrivateKey(keyTag)
		if err != nil {
			return err
		}

		signature, _, err := privateKey.Sign(messageBytes)
		if err != nil {
			return err
		}
		if keyTag.Type() == symbiotic.KeyTypeEcdsaSecp256k1 && len(signature) == 65 {
			signature[64] += 27
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), hexutil.Encode(signature))
		return err
	},
}
