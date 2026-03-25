package tests

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type pruningEntityType string

const entityMeta pruningEntityType = "meta"

type remainingEntity struct {
	Type     pruningEntityType
	Location string
	Key      string
}

type excludedEntityTypes map[pruningEntityType]string

func (e excludedEntityTypes) contains(entityType pruningEntityType) bool {
	_, ok := e[entityType]
	return ok
}

func TestPruningE2E_RemovesAllNonExcludedEntities(t *testing.T) {
	if os.Getenv("E2E_PRUNING_TEST") == "" {
		t.Skip("set E2E_PRUNING_TEST=1 to run pruning e2e tests")
	}

	t.Log("Starting pruning e2e test...")

	ctx := t.Context()
	envInfo := loadEnvInfo(t)
	client := getGRPCClient(t, 0)

	t.Log("Step 1: Finding a committed epoch for pruning...")
	targetEpoch := getCommittedEpochForPruning(t, client)
	t.Logf("Using committed epoch %d", targetEpoch)

	t.Log("Step 2: Creating a real signature request for the target epoch...")
	requestID := createPruningRequest(t, targetEpoch, envInfo)
	t.Logf("Created pruning request %s", requestID)

	t.Log("Step 3: Waiting for signatures and aggregation proof...")
	waitForRequestSignatures(t, client, requestID, len(envInfo.GetSidecarConfigs()))
	waitForRequestProof(t, client, requestID)

	maxRetention := max(
		readUint64Env("RETENTION_VALSET_EPOCHS", 2),
		readUint64Env("RETENTION_SIGNATURE_EPOCHS", 2),
		readUint64Env("RETENTION_PROOF_EPOCHS", 2),
	)
	require.Positive(t, maxRetention, "pruning test requires positive retention")

	t.Log("Step 4: Waiting until the target epoch is eligible for pruning...")
	pruneReadyEpoch := targetEpoch + symbiotic.Epoch(maxRetention)
	require.NoError(t, waitForAPIEpoch(ctx, client, pruneReadyEpoch, 3*waitEpochTimeout()))
	t.Logf("Epoch %d is now eligible for pruning", targetEpoch)

	excluded := excludedEntityTypes{
		entityMeta: "meta stores global pointers rather than pruned epoch entities",
	}
	debugScan := os.Getenv("PRUNING_SCAN_DEBUG") != ""
	var before, finalAfter []remainingEntity
	if debugScan {
		var err error
		before, err = scanSidecarStorage(0, targetEpoch, common.HexToHash(requestID), nil)
		require.NoError(t, err)
	}

	t.Log("Step 5: Waiting for pruning and verifying that no non-excluded entities remain...")
	require.NoError(t, waitForErrorIsNil(ctx, pruningTimeout(), func() error {
		remaining, err := scanSidecarStorage(0, targetEpoch, common.HexToHash(requestID), excluded)
		if err != nil {
			return err
		}
		if debugScan {
			after, scanErr := scanSidecarStorage(0, targetEpoch, common.HexToHash(requestID), nil)
			if scanErr != nil {
				return scanErr
			}
			finalAfter = after
		}
		if len(remaining) > 0 {
			return errors.Errorf("found unpruned entities:\n%s", formatRemainingEntities(remaining))
		}
		return nil
	}))

	if debugScan {
		t.Logf("before pruning:\n%s", formatRemainingEntities(before))
		t.Logf("after pruning:\n%s", formatRemainingEntities(finalAfter))
		t.Logf("pruned entities:\n%s", formatRemainingEntities(diffRemainingEntities(before, finalAfter)))
	}

	t.Log("Pruning e2e test completed successfully")
}

func getCommittedEpochForPruning(t *testing.T, client *apiv1.SymbioticClient) symbiotic.Epoch {
	t.Helper()

	resp, err := client.GetLastAllCommitted(t.Context(), &apiv1.GetLastAllCommittedRequest{})
	require.NoError(t, err)

	var minEpoch uint64
	first := true
	for _, info := range resp.GetEpochInfos() {
		epoch := info.GetLastCommittedEpoch()
		if first || epoch < minEpoch {
			minEpoch = epoch
			first = false
		}
	}
	require.False(t, first, "expected at least one committed epoch")

	return symbiotic.Epoch(minEpoch)
}

func createPruningRequest(t *testing.T, epoch symbiotic.Epoch, envInfo EnvInfo) string {
	t.Helper()

	msg := fmt.Sprintf("pruning-e2e-%d", time.Now().UnixNano())
	var requestID string

	for i := range envInfo.GetSidecarConfigs() {
		client := getGRPCClient(t, i)
		resp, err := client.SignMessage(t.Context(), &apiv1.SignMessageRequest{
			KeyTag:        15,
			Message:       []byte(msg),
			RequiredEpoch: (*uint64)(&epoch),
		})
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

func waitForAPIEpoch(
	ctx context.Context,
	client *apiv1.SymbioticClient,
	targetEpoch symbiotic.Epoch,
	timeout time.Duration,
) error {
	return waitForErrorIsNil(ctx, timeout, func() error {
		resp, err := client.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
		if err != nil {
			return err
		}
		if symbiotic.Epoch(resp.GetEpoch()) < targetEpoch {
			return errors.Errorf("current epoch %d is below target epoch %d", resp.GetEpoch(), targetEpoch)
		}
		return nil
	})
}

func pruningTimeout() time.Duration {
	base := waitEpochTimeout()
	interval := readDurationEnv("PRUNER_INTERVAL", time.Minute)
	return base + 4*interval
}

func scanSidecarStorage(
	sidecarIndex int,
	targetEpoch symbiotic.Epoch,
	requestID common.Hash,
	excluded excludedEntityTypes,
) ([]remainingEntity, error) {
	if strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger") {
		return scanBadgerStorage(sidecarIndex, targetEpoch, requestID, excluded)
	}
	return scanBboltStorage(sidecarIndex, targetEpoch, requestID, excluded)
}

func scanBboltStorage(
	sidecarIndex int,
	targetEpoch symbiotic.Epoch,
	requestID common.Hash,
	excluded excludedEntityTypes,
) ([]remainingEntity, error) {
	dbPath := filepath.Join("..", "temp-network", sidecarStorageDir(sidecarIndex), "relay.db")
	db, err := bolt.Open(dbPath, 0o600, &bolt.Options{ReadOnly: true, Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	epochBytes := bboltEpochBytes(uint64(targetEpoch))
	requestIDBytes := requestID.Bytes()
	requestIDHex := requestID.Hex()
	var remaining []remainingEntity

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			entityType := pruningEntityType(name)
			cursor := bucket.Cursor()
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				if !matchesPrunedEntity(k, v, epochBytes, requestIDBytes, requestIDHex) {
					continue
				}
				if excluded.contains(entityType) {
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
	targetEpoch symbiotic.Epoch,
	requestID common.Hash,
	excluded excludedEntityTypes,
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

	epochBytes := targetEpoch.Bytes()
	requestIDBytes := requestID.Bytes()
	requestIDHex := requestID.Hex()
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
			if !matchesPrunedEntity(key, value, epochBytes, requestIDBytes, requestIDHex) {
				continue
			}
			entityType := badgerEntityType(key)
			if excluded.contains(entityType) {
				continue
			}
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

func matchesPrunedEntity(key, value, epochBytes, requestIDBytes []byte, requestIDHex string) bool {
	if bytes.Contains(key, requestIDBytes) || bytes.Contains(value, requestIDBytes) {
		return true
	}
	if bytes.Contains(key, []byte(requestIDHex)) || bytes.Contains(value, []byte(requestIDHex)) {
		return true
	}
	if bytes.Contains(key, epochBytes) || bytes.Contains(value, epochBytes) {
		return true
	}
	return false
}

func sidecarStorageDir(sidecarIndex int) string {
	return fmt.Sprintf("data-%02d", sidecarIndex+1)
}

func bboltEpochBytes(epoch uint64) []byte {
	b := make([]byte, 8)
	b[0] = byte(epoch >> 56)
	b[1] = byte(epoch >> 48)
	b[2] = byte(epoch >> 40)
	b[3] = byte(epoch >> 32)
	b[4] = byte(epoch >> 24)
	b[5] = byte(epoch >> 16)
	b[6] = byte(epoch >> 8)
	b[7] = byte(epoch)
	return b
}

func readUint64Env(name string, fallback uint64) uint64 {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return fallback
	}
	return parsed
}

func readDurationEnv(name string, fallback time.Duration) time.Duration {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
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

func diffRemainingEntities(before, after []remainingEntity) []remainingEntity {
	afterSet := make(map[string]struct{}, len(after))
	for _, entity := range after {
		afterSet[remainingEntityKey(entity)] = struct{}{}
	}

	diff := make([]remainingEntity, 0, len(before))
	for _, entity := range before {
		if _, ok := afterSet[remainingEntityKey(entity)]; ok {
			continue
		}
		diff = append(diff, entity)
	}
	return diff
}

func remainingEntityKey(entity remainingEntity) string {
	return fmt.Sprintf("%s|%s|%s", entity.Type, entity.Location, entity.Key)
}
