package loadtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	badgerrepo "github.com/symbioticfi/relay/internal/client/repository/badger"
	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"
	"github.com/symbioticfi/relay/internal/client/repository/cached"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type loadConfig struct {
	validators      int
	requestInterval time.Duration
	duration        time.Duration
}

func getLoadConfig() loadConfig {
	cfg := loadConfig{
		validators:      10,
		requestInterval: 100 * time.Millisecond,
		duration:        10 * time.Second,
	}

	if v := os.Getenv("LOADTEST_VALIDATORS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.validators = n
		}
	}
	if v := os.Getenv("LOADTEST_REQUEST_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.requestInterval = d
		}
	}
	if v := os.Getenv("LOADTEST_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.duration = d
		}
	}

	return cfg
}

type testBackend struct {
	repo       cached.Repository
	validators []symbiotic.Validator
	keys       []crypto.PrivateKey
	cleanup    func()
}

func setupBbolt(b *testing.B, numValidators int) testBackend {
	b.Helper()
	dir := b.TempDir()
	repo, err := bboltrepo.New(bboltrepo.Config{
		Dir:     dir,
		Metrics: repoutil.DoNothingMetrics{},
	})
	if err != nil {
		b.Fatalf("failed to create bbolt repo: %v", err)
	}

	backend := testBackend{
		repo:    repo,
		cleanup: func() { repo.Close() },
	}
	backend.keys, backend.validators = generateValidators(b, numValidators)
	seedData(b, repo, backend.validators, backend.keys)
	return backend
}

func setupBadger(b *testing.B, numValidators int) testBackend {
	b.Helper()
	dir := b.TempDir()
	repo, err := badgerrepo.New(badgerrepo.Config{
		Dir:            dir,
		Metrics:        repoutil.DoNothingMetrics{},
		BlockCacheSize: 64 << 20, // 64MB
	})
	if err != nil {
		b.Fatalf("failed to create badger repo: %v", err)
	}

	backend := testBackend{
		repo:    repo,
		cleanup: func() { repo.Close() },
	}
	backend.keys, backend.validators = generateValidators(b, numValidators)
	seedData(b, repo, backend.validators, backend.keys)
	return backend
}

func generateValidators(b *testing.B, n int) ([]crypto.PrivateKey, []symbiotic.Validator) {
	b.Helper()
	keys := make([]crypto.PrivateKey, n)
	validators := make([]symbiotic.Validator, n)

	for i := range n {
		priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		if err != nil {
			b.Fatalf("failed to generate key %d: %v", i, err)
		}
		keys[i] = priv

		opBytes := make([]byte, 20)
		// Use deterministic operator addresses so they sort consistently
		big.NewInt(int64(i + 1)).FillBytes(opBytes)

		validators[i] = symbiotic.Validator{
			Operator:    common.BytesToAddress(opBytes),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{{
				Tag:     symbiotic.KeyTag(15),
				Payload: priv.PublicKey().OnChain(),
			}},
			Vaults: []symbiotic.ValidatorVault{{
				ChainID:     1,
				Vault:       common.BytesToAddress(opBytes),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			}},
		}
	}

	// Sort by operator address ascending (required by saveValidatorSet)
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Operator.Hex() < validators[j].Operator.Hex()
	})
	// Re-order keys to match sorted validators
	keyMap := make(map[common.Address]crypto.PrivateKey, n)
	for i, v := range validators {
		_ = i
		keyMap[v.Operator] = findKeyForValidator(keys, v, validators)
	}

	// Rebuild keys array matching sorted order
	sortedKeys := make([]crypto.PrivateKey, n)
	for i, v := range validators {
		sortedKeys[i] = keyMap[v.Operator]
	}

	return sortedKeys, validators
}

func findKeyForValidator(keys []crypto.PrivateKey, v symbiotic.Validator, _ []symbiotic.Validator) crypto.PrivateKey {
	for _, k := range keys {
		if len(v.Keys) > 0 && common.Bytes2Hex(k.PublicKey().OnChain()) == common.Bytes2Hex(v.Keys[0].Payload) {
			return k
		}
	}
	panic("key not found for validator")
}

func seedData(b *testing.B, repo cached.Repository, validators []symbiotic.Validator, _ []crypto.PrivateKey) {
	b.Helper()
	ctx := context.Background()

	epoch := symbiotic.Epoch(1)

	networkConfig := symbiotic.NetworkConfig{
		VotingPowerProviders:    []symbiotic.CrossChainAddress{{ChainId: 1, Address: common.HexToAddress("0x1")}},
		KeysProvider:            symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x2")},
		Settlements:             []symbiotic.CrossChainAddress{{ChainId: 1, Address: common.HexToAddress("0x3")}},
		VerificationType:        1,
		MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(100000)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(1)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(1000)),
		RequiredKeyTags:         []symbiotic.KeyTag{15},
		RequiredHeaderKeyTag:    15,
		QuorumThresholds:        []symbiotic.QuorumThreshold{{KeyTag: 15, QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(50))}},
		EpochDuration:           600,
		NumAggregators:          1,
		NumCommitters:           1,
		CommitterSlotDuration:   60,
	}

	prevValset := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            symbiotic.Epoch(0),
		CaptureTimestamp: 1000000,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(500)),
		Validators:       validators,
		Status:           symbiotic.HeaderDerived,
	}

	nextValset := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1000000,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(500)),
		Validators:       validators,
		Status:           symbiotic.HeaderDerived,
	}

	message := randomBytes(b)
	sigReq := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       message,
	}

	requestID := common.BytesToHash(randomBytes(b))

	data := entity.NextValsetData{
		PrevValidatorSet:  prevValset,
		PrevNetworkConfig: networkConfig,
		NextValidatorSet:  nextValset,
		NextNetworkConfig: networkConfig,
		SignatureRequest:  &sigReq,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{
			RequestID:      requestID,
			Epoch:          epoch,
			CommitmentData: randomBytes(b),
		},
	}

	if err := repo.SaveNextValsetData(ctx, data); err != nil {
		b.Fatalf("failed to seed data: %v", err)
	}
}

type pregenSignatures struct {
	signatures [][]symbiotic.Signature // [requestIdx][validatorIdx]
	messages   [][]byte
}

func pregenerate(b *testing.B, keys []crypto.PrivateKey, numRequests int) pregenSignatures {
	b.Helper()

	epoch := symbiotic.Epoch(1)
	keyTag := symbiotic.KeyTag(15)

	pg := pregenSignatures{
		signatures: make([][]symbiotic.Signature, numRequests),
		messages:   make([][]byte, numRequests),
	}

	for r := range numRequests {
		message := randomBytes(b)
		pg.messages[r] = message
		pg.signatures[r] = make([]symbiotic.Signature, len(keys))

		for v, priv := range keys {
			rawSig, messageHash, err := priv.Sign(message)
			if err != nil {
				b.Fatalf("failed to sign: %v", err)
			}
			pg.signatures[r][v] = symbiotic.Signature{
				KeyTag:      keyTag,
				Epoch:       epoch,
				MessageHash: messageHash,
				Signature:   rawSig,
				PublicKey:   priv.PublicKey(),
			}
		}
	}

	return pg
}

type latencyCollector struct {
	samples []time.Duration
	idx     atomic.Int64
}

func newLatencyCollector(capacity int) *latencyCollector {
	return &latencyCollector{
		samples: make([]time.Duration, capacity),
	}
}

func (lc *latencyCollector) record(d time.Duration) {
	i := lc.idx.Add(1) - 1
	if int(i) < len(lc.samples) {
		lc.samples[i] = d
	}
}

func (lc *latencyCollector) results() []time.Duration {
	n := int(lc.idx.Load())
	if n > len(lc.samples) {
		n = len(lc.samples)
	}
	res := make([]time.Duration, n)
	copy(res, lc.samples[:n])
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func BenchmarkSaveSignature(b *testing.B) {
	cfg := getLoadConfig()
	numRequests := int(cfg.duration / cfg.requestInterval)
	if numRequests < 1 {
		numRequests = 1
	}

	b.Logf("Config: validators=%d, requestInterval=%s, duration=%s, totalRequests=%d",
		cfg.validators, cfg.requestInterval, cfg.duration, numRequests)

	b.Run("bbolt", func(b *testing.B) {
		backend := setupBbolt(b, cfg.validators)
		defer backend.cleanup()
		runLoadTest(b, backend, cfg, numRequests)
	})

	b.Run("badger", func(b *testing.B) {
		backend := setupBadger(b, cfg.validators)
		defer backend.cleanup()
		runLoadTest(b, backend, cfg, numRequests)
	})
}

func runLoadTest(b *testing.B, backend testBackend, cfg loadConfig, numRequests int) {
	b.Helper()

	b.Logf("Pre-generating %d requests × %d validators = %d signatures...",
		numRequests, cfg.validators, numRequests*cfg.validators)

	pg := pregenerate(b, backend.keys, numRequests)

	totalOps := numRequests * cfg.validators
	collector := newLatencyCollector(totalOps)
	var errCount atomic.Int64

	ctx := context.Background()

	b.ResetTimer()

	for i := range b.N {
		_ = i
		var wg sync.WaitGroup
		start := time.Now()

		ticker := time.NewTicker(cfg.requestInterval)

		for r := range numRequests {
			if r > 0 {
				<-ticker.C
			}

			wg.Add(cfg.validators)
			for v := range cfg.validators {
				go func() {
					defer wg.Done()
					sig := pg.signatures[r][v]
					opStart := time.Now()
					err := backend.repo.SaveSignature(ctx, sig, backend.validators[v], uint32(v))
					elapsed := time.Since(opStart)
					if err != nil {
						errCount.Add(1)
					} else {
						collector.record(elapsed)
					}
				}()
			}
		}

		wg.Wait()
		ticker.Stop()
		elapsed := time.Since(start)

		sorted := collector.results()
		successOps := len(sorted)
		errors := errCount.Load()
		opsPerSec := float64(successOps) / elapsed.Seconds()

		b.ReportMetric(opsPerSec, "ops/sec")
		b.ReportMetric(float64(percentile(sorted, 0.50).Microseconds())/1000, "p50_ms")
		b.ReportMetric(float64(percentile(sorted, 0.95).Microseconds())/1000, "p95_ms")
		b.ReportMetric(float64(percentile(sorted, 0.99).Microseconds())/1000, "p99_ms")
		b.ReportMetric(float64(errors), "errors")

		fmt.Printf("\n  %-8s ops=%-6d ops/sec=%-10.1f errors=%-4d duration=%s\n",
			"",
			successOps, opsPerSec, errors, elapsed.Round(time.Millisecond))
		fmt.Printf("  %-8s p50=%-10s p95=%-10s p99=%-10s max=%s\n",
			"",
			percentile(sorted, 0.50).Round(time.Microsecond),
			percentile(sorted, 0.95).Round(time.Microsecond),
			percentile(sorted, 0.99).Round(time.Microsecond),
			percentile(sorted, 1.0).Round(time.Microsecond))
	}
}

func randomBytes(b *testing.B) []byte {
	b.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		b.Fatalf("failed to generate random bytes: %v", err)
	}
	return buf
}
