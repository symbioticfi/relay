package keys

import (
	"strconv"

	"github.com/pterm/pterm"
	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"

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

		tableData := pterm.TableData{{"#", "Alias", "Public Key"}}
		for i, alias := range aliases {
			ns, kType, id, err := keyprovider.AliasToKeyTypeId(alias)
			if err != nil {
				return err
			}
			pk, err := keyStore.GetPrivateKeyByNamespaceTypeId(ns, kType, id)
			if err != nil {
				return err
			}
			prettyPk, err := pk.PublicKey().OnChain().MarshalText()
			if err != nil {
				return err
			}
			tableData = append(tableData, []string{
				strconv.Itoa(i + 1),
				alias,
				string(prettyPk),
			})
		}
		return pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}
