package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show badger store information",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return runInfo(ctx)
	},
}

func runInfo(ctx context.Context) error {
	opts := badger.DefaultOptions(globalFlags.StorePath).WithReadOnly(true)
	db, err := badger.Open(opts)
	if err != nil {
		pterm.Error.Printf("Failed to open badger store: %v\n", err)
		return err
	}
	defer db.Close()

	var keyCount int
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		if !infoFlags.Full {
			opts.PrefetchValues = false
		}

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			keyCount++

			if infoFlags.Keys || infoFlags.Full {
				key := item.Key()
				pterm.Printf("Key: %s", string(key))

				if infoFlags.Full {
					err := item.Value(func(val []byte) error {
						pterm.Printf(" | Value: %s", string(val))
						return nil
					})
					if err != nil {
						pterm.Printf(" | Value: <error reading value: %v>", err)
					}
				}
				pterm.Println()
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	pterm.Success.Printf("Total keys: %d\n", keyCount)
	return nil
}
