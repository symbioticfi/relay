package keys

import (
	"github.com/spf13/cobra"
	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
)

var migrateKeysCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Atomically encrypt all legacy keystore entries with the store password",
	RunE: func(cmd *cobra.Command, args []string) error {
		password := globalFlags.Password
		var err error
		if password == "" {
			password, err = cmdhelpers.GetPassword()
			if err != nil {
				return err
			}
		}
		provider, err := keyprovider.NewKeystoreProvider(globalFlags.Path, password)
		if err != nil {
			return err
		}
		return provider.Migrate(password)
	},
}
