package badger

import (
	"github.com/spf13/cobra"
)

func NewBadgerCmd() *cobra.Command {
	badgerCmd.AddCommand(infoCmd)

	initFlags()

	return badgerCmd
}

var badgerCmd = &cobra.Command{
	Use:   "badger",
	Short: "Badger store utility tool",
}

type GlobalFlags struct {
	StorePath string
}

type InfoFlags struct {
	Keys bool
	Full bool
}

var globalFlags GlobalFlags
var infoFlags InfoFlags

func initFlags() {
	badgerCmd.PersistentFlags().StringVarP(&globalFlags.StorePath, "store-path", "s", "./badger-store", "Path to badger store")

	infoCmd.Flags().BoolVarP(&infoFlags.Keys, "keys", "k", false, "List all keys in the store")
	infoCmd.Flags().BoolVarP(&infoFlags.Full, "full", "f", false, "Show full key-value pairs")
}
