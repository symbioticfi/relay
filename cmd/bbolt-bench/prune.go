package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type PruneConfig struct {
	DBPath         string
	NoFreelistSync bool
	KeepEpochs     int
	BatchSize      int
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Run pruning on the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := PruneConfig{
			DBPath:         flagDB,
			NoFreelistSync: flagNoFreelistSync,
		}
		cfg.KeepEpochs, _ = cmd.Flags().GetInt("keep-epochs")
		cfg.BatchSize, _ = cmd.Flags().GetInt("batch-size")
		return runPrune(cfg)
	},
}

func init() {
	pruneCmd.Flags().Int("keep-epochs", 5, "number of recent epochs to keep")
	pruneCmd.Flags().Int("batch-size", 100, "requestIDs per prune transaction")
}

func runPrune(cfg PruneConfig) error {
	repo, err := openRepo(cfg.DBPath, cfg.NoFreelistSync)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}
	defer repo.Close()

	ctx := context.Background()

	latestEpoch, err := repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return fmt.Errorf("get latest epoch: %w", err)
	}

	oldestEpoch, err := repo.GetOldestValidatorSetEpoch(ctx)
	if err != nil {
		return fmt.Errorf("get oldest epoch: %w", err)
	}

	oldestToKeep := latestEpoch - symbiotic.Epoch(cfg.KeepEpochs)
	if oldestToKeep <= oldestEpoch {
		fmt.Printf("Nothing to prune: oldest=%d, latest=%d, keep=%d\n",
			oldestEpoch, latestEpoch, cfg.KeepEpochs)
		return nil
	}

	sizeBefore := fileSize(cfg.DBPath)
	totalStart := time.Now()
	pruned := 0

	fmt.Printf("Pruning epochs %d..%d (keeping %d..%d)\n",
		oldestEpoch, oldestToKeep-1, oldestToKeep, latestEpoch)

	totalEpochs := int(oldestToKeep - oldestEpoch)
	for epoch := oldestEpoch; epoch < oldestToKeep; epoch++ {
		epochStart := time.Now()
		epochNum := int(epoch-oldestEpoch) + 1

		fmt.Printf("  Epoch %d (%d/%d): pruning proofs...\n", epoch, epochNum, totalEpochs)
		if err := repo.PruneProofEntities(ctx, epoch, cfg.BatchSize); err != nil {
			return fmt.Errorf("prune proof entities epoch %d: %w", epoch, err)
		}
		fmt.Printf("  Epoch %d (%d/%d): pruning signatures...\n", epoch, epochNum, totalEpochs)
		if err := repo.PruneSignatureEntitiesForEpoch(ctx, epoch, cfg.BatchSize); err != nil {
			return fmt.Errorf("prune signature entities epoch %d: %w", epoch, err)
		}
		fmt.Printf("  Epoch %d (%d/%d): pruning valset...\n", epoch, epochNum, totalEpochs)
		if err := repo.PruneValsetEntities(ctx, epoch, cfg.BatchSize); err != nil {
			return fmt.Errorf("prune valset entities epoch %d: %w", epoch, err)
		}
		fmt.Printf("  Epoch %d (%d/%d): pruning request indices...\n", epoch, epochNum, totalEpochs)
		if err := repo.PruneRequestIDEpochIndices(ctx, epoch, cfg.BatchSize); err != nil {
			return fmt.Errorf("prune request id epoch indices epoch %d: %w", epoch, err)
		}

		pruned++
		elapsed := time.Since(totalStart)
		etaPerEpoch := elapsed / time.Duration(epochNum)
		remaining := etaPerEpoch * time.Duration(totalEpochs-epochNum)
		fmt.Printf("  Epoch %d done in %s, ETA %s\n",
			epoch, time.Since(epochStart).Round(time.Millisecond), remaining.Round(time.Second))
	}

	sizeAfter := fileSize(cfg.DBPath)
	fmt.Printf("\nPrune complete: %d epochs in %s\n", pruned, time.Since(totalStart).Round(time.Second))
	fmt.Printf("File size: %s -> %s (file doesn't shrink, freed pages go to freelist)\n",
		formatBytes(sizeBefore), formatBytes(sizeAfter))
	printDBStats(cfg.DBPath, repo.Stats())
	return nil
}
