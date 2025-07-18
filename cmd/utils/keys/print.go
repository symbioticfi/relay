package keys

import (
	"fmt"

	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"

	"github.com/spf13/cobra"
)

var printKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all keys",
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

		aliases := keyStore.GetAliases()

		fmt.Printf("\nKeys (%d):\n", len(aliases)) // assuming 'keys' is your collection
		fmt.Printf("   # | Alias                | Public Key\n")

		for i, alias := range aliases {
			keyTag, err := keyprovider.AliasToKeyTag(alias)
			if err != nil {
				return err
			}

			pk, err := keyStore.GetPrivateKey(keyTag)
			if err != nil {
				return err
			}

			prettyPk, err := pk.PublicKey().OnChain().MarshalText()
			if err != nil {
				return err
			}

			fmt.Printf("   %d | %-20s | %s\n", i+1, alias, prettyPk)
		}

		return nil
	},
}
