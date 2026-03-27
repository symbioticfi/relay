package tests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type pruningEntityType string

type remainingEntity struct {
	Type     pruningEntityType
	Location string
	Key      string
}

type pruningScope struct {
	epochBytes     []byte
	requestIDBytes [][]byte
	requestIDHexes [][]byte
}

func TestPruningE2E_RemovesAllNonExcludedEntities(t *testing.T) {
	t.Log("Starting pruning e2e test...")

	ctx := t.Context()
	envInfo := loadEnvInfo(t)
	sidecarConfig := loadSidecarConfig(t)
	client := getGRPCClient(t, 0)
	scanSidecarIndex := 0
	if strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger") {
		scanSidecarIndex = int(envInfo.Operators)
	}

	t.Log("Step 1: Finding a committed epoch for pruning...")
	targetEpoch := getCommittedEpochForPruning(t, client)
	t.Logf("Using committed epoch %d", targetEpoch)

	t.Log("Step 2: Creating a real signing request for the target epoch...")
	requestID := createSignatureRequestForEpoch(t, targetEpoch, envInfo)
	t.Logf("Created signing request %s", requestID)

	t.Log("Step 3: Waiting for signatures and aggregation proof...")
	waitForRequestSignatures(t, client, requestID, len(envInfo.GetSidecarConfigs()))
	waitForRequestProof(t, client, requestID)

	maxRetention := max(
		uint64(sidecarConfig.Retention.ValsetEpochs),
		uint64(sidecarConfig.Retention.SignatureEpochs),
		uint64(sidecarConfig.Retention.ProofEpochs),
	)
	require.Positive(t, maxRetention, "pruning test requires positive retention")

	t.Log("Step 4: Waiting until the target epoch is eligible for pruning...")
	pruneReadyEpoch := targetEpoch + symbiotic.Epoch(maxRetention)
	require.NoError(t, waitForErrorIsNil(ctx, 3*waitEpochTimeout(), func() error {
		resp, err := client.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
		if err != nil {
			return err
		}
		if symbiotic.Epoch(resp.GetEpoch()) < pruneReadyEpoch {
			return errors.Errorf("current epoch %d is below target epoch %d", resp.GetEpoch(), pruneReadyEpoch)
		}
		return nil
	}))
	t.Logf("Epoch %d is now eligible for pruning", targetEpoch)

	scope := buildPruningScope(targetEpoch, common.HexToHash(requestID))
	var lastRemaining []remainingEntity

	t.Log("Step 5: Waiting for pruning and verifying that no non-excluded entities remain...")
	if strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger") {
		time.Sleep(badgerOfflineScanDelay(t))
		require.NoError(t, stopSidecarForStorageScan(t, scanSidecarIndex, envInfo))
		defer func() {
			require.NoError(t, startSidecarAfterStorageScan(t, scanSidecarIndex, envInfo))
		}()

		remaining, err := scanSidecarStorage(scanSidecarIndex, scope)
		require.NoError(t, err)
		require.Emptyf(t, remaining, "found unpruned entities:\n%s", formatRemainingEntities(remaining))
		t.Log("Pruning e2e test completed successfully")
		return
	}

	err := waitForErrorIsNil(ctx, pruningTimeout(t), func() error {
		remaining, err := scanSidecarStorage(scanSidecarIndex, scope)
		if err != nil {
			return err
		}
		lastRemaining = remaining
		if len(remaining) > 0 {
			return errors.Errorf("found unpruned entities:\n%s", formatRemainingEntities(remaining))
		}
		return nil
	})
	require.NoErrorf(t, err, "last remaining entities before timeout:\n%s", formatRemainingEntities(lastRemaining))

	t.Log("Pruning e2e test completed successfully")
}

func TestMatchesPrunedEntity(t *testing.T) {
	t.Parallel()

	requestID := common.HexToHash("0x4b939f34668a2b051228cf038dd1654aa57dbac80053cd68fee2e9d68eb9c5a6")
	scope := buildPruningScope(symbiotic.Epoch(15), requestID)

	t.Run("matches epoch in key", func(t *testing.T) {
		key := append([]byte("validator:"), append(symbiotic.Epoch(15).Bytes(), []byte(":0xoperator")...)...)
		require.True(t, matchesPrunedEntity(key, []byte("ignored"), scope))
	})

	t.Run("matches request id in key", func(t *testing.T) {
		key := append([]byte("signature:"), requestID.Bytes()...)
		require.True(t, matchesPrunedEntity(key, []byte("ignored"), scope))
	})

	t.Run("does not match unrelated value bytes", func(t *testing.T) {
		key := append([]byte("validator:"), append(symbiotic.Epoch(16).Bytes(), []byte(":0xoperator")...)...)
		value := append([]byte("payload:"), symbiotic.Epoch(15).Bytes()...)
		require.False(t, matchesPrunedEntity(key, value, scope))
	})
}

func stopSidecarForStorageScan(t *testing.T, sidecarIndex int, envInfo EnvInfo) error {
	t.Helper()

	serviceName := storageScanSidecarName(sidecarIndex, envInfo)
	cmd := exec.CommandContext(t.Context(), "docker", "compose", "stop", "-t", "1", serviceName)
	cmd.Dir = filepath.Join("..", "temp-network")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("failed to stop %s: %w: %s", serviceName, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func startSidecarAfterStorageScan(t *testing.T, sidecarIndex int, envInfo EnvInfo) error {
	t.Helper()

	serviceName := storageScanSidecarName(sidecarIndex, envInfo)
	if err := startContainer(t.Context(), serviceName); err != nil {
		return errors.Errorf("failed to start %s: %w", serviceName, err)
	}
	return nil
}

func storageScanSidecarName(sidecarIndex int, envInfo EnvInfo) string {
	if sidecarIndex >= int(envInfo.Operators) {
		return "relay-sidecar-extra"
	}
	return fmt.Sprintf("relay-sidecar-%d", sidecarIndex+1)
}

func getCommittedEpochForPruning(t *testing.T, client *apiv1.SymbioticClient) symbiotic.Epoch {
	t.Helper()

	resp, err := client.GetLastAllCommitted(t.Context(), &apiv1.GetLastAllCommittedRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetEpochInfos(), "expected at least one committed epoch")

	return symbiotic.Epoch(lo.Min(lo.Map(lo.Values(resp.GetEpochInfos()), func(info *apiv1.ChainEpochInfo, _ int) uint64 {
		return info.GetLastCommittedEpoch()
	})))
}

func createSignatureRequestForEpoch(t *testing.T, epoch symbiotic.Epoch, envInfo EnvInfo) string {
	t.Helper()

	msg := fmt.Sprintf("pruning-e2e-%d", time.Now().UnixNano())
	var requestID string

	for i := range envInfo.GetSidecarConfigs() {
		client := getGRPCClient(t, i)
		var (
			resp *apiv1.SignMessageResponse
			err  error
		)

		for attempts := 1; attempts <= 3; attempts++ {
			resp, err = client.SignMessage(t.Context(), &apiv1.SignMessageRequest{
				KeyTag:        15,
				Message:       []byte(msg),
				RequiredEpoch: (*uint64)(&epoch),
			})
			if err == nil {
				break
			}
		}
		require.NoErrorf(t, err, "failed to sign message on sidecar %d", i)
		require.NotEmpty(t, resp.GetRequestId())

		if requestID == "" {
			requestID = resp.GetRequestId()
			continue
		}
		require.Equalf(t, requestID, resp.GetRequestId(), "request id mismatch on sidecar %d", i)
	}

	return requestID
}

func waitForRequestSignatures(t *testing.T, client *apiv1.SymbioticClient, requestID string, signerCount int) {
	t.Helper()

	require.NoError(t, waitForErrorIsNil(t.Context(), waitEpochTimeout(), func() error {
		resp, err := client.GetSignatures(t.Context(), &apiv1.GetSignaturesRequest{RequestId: requestID})
		if err != nil {
			return err
		}
		if len(resp.GetSignatures()) == 0 {
			return errors.Errorf("no signatures available for request %s yet", requestID)
		}
		threshold := signerCount*2/3 + 1
		if len(resp.GetSignatures()) < threshold {
			return errors.Errorf("received %d signatures, need at least %d", len(resp.GetSignatures()), threshold)
		}
		return nil
	}))
}

func waitForRequestProof(t *testing.T, client *apiv1.SymbioticClient, requestID string) {
	t.Helper()

	require.NoError(t, waitForErrorIsNil(t.Context(), 2*waitEpochTimeout(), func() error {
		resp, err := client.GetAggregationProof(t.Context(), &apiv1.GetAggregationProofRequest{
			RequestId: requestID,
		})
		if err != nil {
			return err
		}
		if resp.GetAggregationProof() == nil || len(resp.GetAggregationProof().GetProof()) == 0 {
			return errors.Errorf("aggregation proof for request %s is empty", requestID)
		}
		return nil
	}))
}

func pruningTimeout(t *testing.T) time.Duration {
	t.Helper()

	interval, err := time.ParseDuration(loadSidecarConfig(t).Pruner.Interval)
	require.NoError(t, err)

	return waitEpochTimeout() + 4*interval
}

func badgerOfflineScanDelay(t *testing.T) time.Duration {
	t.Helper()

	interval, err := time.ParseDuration(loadSidecarConfig(t).Pruner.Interval)
	require.NoError(t, err)

	return 4 * interval
}

func scanSidecarStorage(
	sidecarIndex int,
	scope pruningScope,
) ([]remainingEntity, error) {
	if strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger") {
		return scanBadgerStorage(sidecarIndex, scope)
	}
	return scanBboltStorage(sidecarIndex, scope)
}

func scanBboltStorage(
	sidecarIndex int,
	scope pruningScope,
) ([]remainingEntity, error) {
	dbPath := filepath.Join("..", "temp-network", sidecarStorageDir(sidecarIndex), "relay.db")
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{ReadOnly: true, Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var remaining []remainingEntity

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			entityType := pruningEntityType(name)
			cursor := bucket.Cursor()
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				if !matchesPrunedEntity(k, v, scope) {
					continue
				}
				remaining = append(remaining, remainingEntity{
					Type:     entityType,
					Location: string(name),
					Key:      hex.EncodeToString(k),
				})
			}
			return nil
		})
	})
	return remaining, err
}

func scanBadgerStorage(
	sidecarIndex int,
	scope pruningScope,
) ([]remainingEntity, error) {
	dir := filepath.Join("..", "temp-network", sidecarStorageDir(sidecarIndex))
	opts := badger.DefaultOptions(dir).
		WithReadOnly(true).
		WithBypassLockGuard(true).
		WithLogger(nil)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var remaining []remainingEntity

	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if !matchesPrunedEntity(key, value, scope) {
				continue
			}
			entityType := badgerEntityType(key)
			remaining = append(remaining, remainingEntity{
				Type:     entityType,
				Location: "badger",
				Key:      hex.EncodeToString(key),
			})
		}
		return nil
	})

	return remaining, err
}

func badgerEntityType(key []byte) pruningEntityType {
	if bytes.HasPrefix(key, []byte("request_id_epoch")) {
		return "request_id_epoch"
	}
	if idx := bytes.IndexByte(key, ':'); idx > 0 {
		return pruningEntityType(key[:idx])
	}
	return pruningEntityType(key)
}

func buildPruningScope(targetEpoch symbiotic.Epoch, seedRequestID common.Hash) pruningScope {
	scope := pruningScope{
		epochBytes:     targetEpoch.Bytes(),
		requestIDBytes: nil,
		requestIDHexes: nil,
	}
	if seedRequestID != (common.Hash{}) {
		scope.requestIDBytes = [][]byte{seedRequestID.Bytes()}
		scope.requestIDHexes = [][]byte{[]byte(seedRequestID.Hex())}
	}
	return scope
}

func matchesPrunedEntity(key, _ []byte, scope pruningScope) bool {
	for _, requestIDBytes := range scope.requestIDBytes {
		if bytes.Contains(key, requestIDBytes) {
			return true
		}
	}
	for _, requestIDHex := range scope.requestIDHexes {
		if bytes.Contains(key, requestIDHex) {
			return true
		}
	}
	return bytes.Contains(key, scope.epochBytes)
}

func sidecarStorageDir(sidecarIndex int) string {
	return fmt.Sprintf("data-%02d", sidecarIndex+1)
}

func formatRemainingEntities(entities []remainingEntity) string {
	if len(entities) == 0 {
		return "(none)"
	}
	lines := make([]string, 0, len(entities))
	for _, entity := range entities {
		lines = append(lines, fmt.Sprintf("- %s [%s] %s", entity.Type, entity.Location, entity.Key))
	}
	return strings.Join(lines, "\n")
}
