package prune

import (
	"github.com/spf13/cobra"
)

type Flags struct {
	StorageDir           string
	ValsetEpochs         uint64
	ProofEpochs          uint64
	SignatureEpochs      uint64
	Compact              bool
	BadgerFlattenWorkers int
	PruneBatchSize       int
}

var flags Flags

func NewPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune old epoch data from the relay storage (offline)",
		Long: "Offline pruning of valset / proof / signature entities older than the " +
			"configured retention. Optionally compacts the database when --compact is set. " +
			"The sidecar must be stopped (the DB is opened with an exclusive file-lock).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.StorageDir, "storage-dir", ".data", "Directory containing the relay storage (badger or bbolt)")
	cmd.Flags().Uint64Var(&flags.ValsetEpochs, "retention.valset-epochs", 0, "Keep this many most-recent epochs of valset data (0 = skip)")
	cmd.Flags().Uint64Var(&flags.ProofEpochs, "retention.proof-epochs", 0, "Keep this many most-recent epochs of aggregation proofs (0 = skip)")
	cmd.Flags().Uint64Var(&flags.SignatureEpochs, "retention.signature-epochs", 0, "Keep this many most-recent epochs of signatures (0 = skip)")
	cmd.Flags().BoolVar(&flags.Compact, "compact", true, "After pruning, compact the database file (bbolt: rewrite; badger: Flatten + value log GC)")
	cmd.Flags().IntVar(&flags.BadgerFlattenWorkers, "badger.flatten-workers", 4, "Number of parallel workers for badger Flatten (only when --compact is set)")
	cmd.Flags().IntVar(&flags.PruneBatchSize, "prune-batch-size", 1000, "Number of request IDs to delete per database transaction (larger = faster but holds writer lock longer)")

	return cmd
}
