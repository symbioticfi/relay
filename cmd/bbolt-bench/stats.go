package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

func printDBStats(dbPath string, stats bolt.Stats) {
	size := fileSize(dbPath)

	fmt.Printf("\nDB Stats:\n")
	fmt.Printf("  %-18s %s\n", "File size:", formatBytes(size))
	fmt.Printf("  %-18s %-12s (pages on freelist, grows after pruning)\n", "Free pages:", formatInt(stats.FreePageN))
	fmt.Printf("  %-18s %-12s (pages awaiting release by open read txns)\n", "Pending pages:", formatInt(stats.PendingPageN))
	fmt.Printf("  %-18s %-12s (freelist struct size, serialized on each commit when freelist-sync=on)\n", "Freelist inuse:", formatBytes(int64(stats.FreelistInuse)))
	fmt.Printf("  %-18s %-12s (total bytes in free pages = wasted space until compaction)\n", "Free alloc:", formatBytes(int64(stats.FreeAlloc)))
}

func printTxStatsDelta(before, after bolt.TxStats) {
	diff := after.Sub(&before)
	fmt.Printf("  %-18s %-12s (fsync duration, includes freelist write)\n", "Write time:", diff.GetWriteTime())
	fmt.Printf("  %-18s %-12s (B-tree node restructuring, CPU-bound)\n", "Spill time:", diff.GetSpillTime())
	fmt.Printf("  %-18s %-12s (number of disk writes)\n", "Writes:", formatInt64(diff.GetWrite()))
	fmt.Printf("  %-18s %-12s (B-tree rebalancing after deletes)\n", "Rebalance time:", diff.GetRebalanceTime())
}

func printLatencyStats(label string, samples []time.Duration, totalDuration time.Duration, requestCount int, validators ...int) {
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })

	numValidators := 0
	if len(validators) > 0 {
		numValidators = validators[0]
	}
	totalSignatures := requestCount * numValidators

	fmt.Printf("\n=== %s ===\n", label)
	fmt.Printf("Requests: %d, Validators: %d, Total signatures: %d\n", requestCount, numValidators, totalSignatures)
	fmt.Printf("Duration: %s\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("Throughput: %.1f req/sec, %.1f sig/sec\n",
		float64(requestCount)/totalDuration.Seconds(),
		float64(totalSignatures)/totalDuration.Seconds())
	fmt.Printf("Latency (per request): p50=%s  p95=%s  p99=%s  max=%s\n",
		pct(samples, 0.50), pct(samples, 0.95), pct(samples, 0.99), pct(samples, 1.0))
}

func pct(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx].Round(time.Microsecond)
}

type EpochWriteResult struct {
	Epoch     int
	P95       time.Duration
	P50       time.Duration
	FileSize  int64
	Duration  time.Duration
	SigPerSec float64
}

// RunAllResults holds results from one run-all mode (sync on or off).
type RunAllResults struct {
	WriteEmpty            BenchWriteResult
	WriteAfterCompact     BenchWriteResult
	WritePerPopulateEpoch []EpochWriteResult
	WritePerPruneEpoch    []EpochWriteResult
	OpenTime              time.Duration
	OpenTimeCompacted     time.Duration
	CompactDuration       time.Duration
	StatsAfterPopulate    bolt.Stats
	StatsAfterPrune       bolt.Stats
	FileSizeAfterPopulate int64
	FileSizeBeforeCompact int64
	FileSizeAfterCompact  int64
}

func printSummaryTable(on, off RunAllResults) {
	sep := strings.Repeat("═", 72)

	row := func(name string, onVal, offVal string) {
		fmt.Printf("║ %-28s║ %-20s║ %-20s║\n", name, onVal, offVal)
	}
	header := func(name string) {
		fmt.Printf("╠%s╣\n", sep)
		fmt.Printf("║ %-70s║\n", name)
		fmt.Printf("╠%s╣\n", sep)
	}

	fmt.Printf("\n╔%s╗\n", sep)
	fmt.Printf("║ %-28s║ %-20s║ %-20s║\n", "Metric", "sync=ON", "sync=OFF")

	header("WRITE PER POPULATE EPOCH")
	fmt.Printf("║ %-28s║ %-20s║ %-20s║\n", "", "p95 / sig/s / size", "p95 / sig/s / size")
	row("empty DB",
		fmt.Sprintf("%s/%.0f", pct(on.WriteEmpty.Samples, 0.95), on.WriteEmpty.SigPerSec()),
		fmt.Sprintf("%s/%.0f", pct(off.WriteEmpty.Samples, 0.95), off.WriteEmpty.SigPerSec()))
	for i := range on.WritePerPopulateEpoch {
		onE := on.WritePerPopulateEpoch[i]
		offE := off.WritePerPopulateEpoch[i]
		row(fmt.Sprintf("after epoch %d", onE.Epoch),
			fmt.Sprintf("%s/%.0f/%s", onE.P95, onE.SigPerSec, formatBytes(onE.FileSize)),
			fmt.Sprintf("%s/%.0f/%s", offE.P95, offE.SigPerSec, formatBytes(offE.FileSize)))
	}

	header("WRITE PER PRUNED EPOCH")
	fmt.Printf("║ %-28s║ %-20s║ %-20s║\n", "", "p95 / sig/s / size", "p95 / sig/s / size")
	for i := range on.WritePerPruneEpoch {
		onE := on.WritePerPruneEpoch[i]
		offE := off.WritePerPruneEpoch[i]
		row(fmt.Sprintf("after pruning epoch %d", onE.Epoch),
			fmt.Sprintf("%s/%.0f/%s", onE.P95, onE.SigPerSec, formatBytes(onE.FileSize)),
			fmt.Sprintf("%s/%.0f/%s", offE.P95, offE.SigPerSec, formatBytes(offE.FileSize)))
	}
	row("after compact",
		fmt.Sprintf("%s/%.0f", pct(on.WriteAfterCompact.Samples, 0.95), on.WriteAfterCompact.SigPerSec()),
		fmt.Sprintf("%s/%.0f", pct(off.WriteAfterCompact.Samples, 0.95), off.WriteAfterCompact.SigPerSec()))

	header("OPEN / COMPACT")
	row("Open time (after prune)", on.OpenTime.Round(time.Millisecond).String(), off.OpenTime.Round(time.Millisecond).String())
	row("Open time (after compact)", on.OpenTimeCompacted.Round(time.Millisecond).String(), off.OpenTimeCompacted.Round(time.Millisecond).String())
	row("Compact duration", on.CompactDuration.Round(time.Millisecond).String(), off.CompactDuration.Round(time.Millisecond).String())

	header("DB STATE AFTER POPULATE")
	row("File size", formatBytes(on.FileSizeAfterPopulate), formatBytes(off.FileSizeAfterPopulate))
	row("Free pages", formatInt(on.StatsAfterPopulate.FreePageN), formatInt(off.StatsAfterPopulate.FreePageN))
	row("Freelist inuse", formatBytes(int64(on.StatsAfterPopulate.FreelistInuse)), formatBytes(int64(off.StatsAfterPopulate.FreelistInuse)))
	row("Free alloc", formatBytes(int64(on.StatsAfterPopulate.FreeAlloc)), formatBytes(int64(off.StatsAfterPopulate.FreeAlloc)))

	header("DB STATE AFTER PRUNE")
	row("File size", formatBytes(on.FileSizeBeforeCompact), formatBytes(off.FileSizeBeforeCompact))
	row("Free pages", formatInt(on.StatsAfterPrune.FreePageN), formatInt(off.StatsAfterPrune.FreePageN))
	row("Freelist inuse", formatBytes(int64(on.StatsAfterPrune.FreelistInuse)), formatBytes(int64(off.StatsAfterPrune.FreelistInuse)))
	row("Free alloc", formatBytes(int64(on.StatsAfterPrune.FreeAlloc)), formatBytes(int64(off.StatsAfterPrune.FreeAlloc)))

	header("DB STATE AFTER COMPACT")
	row("File size", formatBytes(on.FileSizeAfterCompact), formatBytes(off.FileSizeAfterCompact))

	fmt.Printf("╚%s╝\n", sep)
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func formatInt64(n int64) string {
	return formatInt(int(n))
}
