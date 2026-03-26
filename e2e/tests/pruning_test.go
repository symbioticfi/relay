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

func TestPruningE2E_RemovesAllNonExcludedEntities(t *testing.T) {
	t.Log("Starting pruning e2e test...")

	ctx := t.Context()
	envInfo := loadEnvInfo(t)
	client := getGRPCClient(t, 0)

	t.Log("Step 1: Finding a committed epoch for pruning...")
	targetEpoch := getCommittedEpochForPruning(t, client)
	t.Logf("Using committed epoch %d", targetEpoch)

	t.Log("Step 2: Creating a real signing request for the target epoch...")
	requestID := createSignatureRequestForEpoch(t, targetEpoch, envInfo)
	t.Logf("Created signing request %s", requestID)

	t.Log("Step 3: Waiting for signatures and aggregation proof...")
	waitForRequestSignatures(t, client, requestID, len(envInfo.GetSidecarConfigs()))
	waitForRequestProof(t, client, requestID)

	retentionValsetEpochs := uint64(readPositiveIntEnv("RETENTION_VALSET_EPOCHS", 2))
	retentionSignatureEpochs := uint64(readPositiveIntEnv("RETENTION_SIGNATURE_EPOCHS", 2))
	retentionProofEpochs := uint64(readPositiveIntEnv("RETENTION_PROOF_EPOCHS", 2))
	useBadgerStorage := strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger")

	maxRetention := max(
		retentionValsetEpochs,
		retentionSignatureEpochs,
		retentionProofEpochs,
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

	scanSidecarIndex := 0
	if useBadgerStorage {
		scanSidecarIndex = len(envInfo.GetSidecarConfigs()) - 1
	}

	debugScan := os.Getenv("PRUNING_SCAN_DEBUG") != ""
	var before, finalAfter []remainingEntity
	if debugScan && !useBadgerStorage {
		var err error
		before, err = scanSidecarStorage(scanSidecarIndex, targetEpoch, common.HexToHash(requestID))
		require.NoError(t, err)
	}

	t.Log("Step 5: Waiting for pruning and verifying that no non-excluded entities remain...")
	if useBadgerStorage {
		t.Log("Waiting for a few pruner intervals before scanning badger storage offline...")
		time.Sleep(badgerOfflineScanDelay())

		t.Log("Stopping one sidecar to scan badger storage offline...")
		stopSidecarForStorageScan(t, scanSidecarIndex)

		remaining, err := scanSidecarStorage(scanSidecarIndex, targetEpoch, common.HexToHash(requestID))
		if err != nil {
			t.Fatal(err)
		}
		if debugScan {
			finalAfter = remaining
		}
		if len(remaining) > 0 {
			t.Fatalf("found unpruned entities:\n%s", formatRemainingEntities(remaining))
		}
	} else {
		require.NoError(t, waitForErrorIsNil(ctx, pruningTimeout(), func() error {
			remaining, err := scanSidecarStorage(scanSidecarIndex, targetEpoch, common.HexToHash(requestID))
			if err != nil {
				return err
			}
			if debugScan {
				after, scanErr := scanSidecarStorage(scanSidecarIndex, targetEpoch, common.HexToHash(requestID))
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
	}

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

func pruningTimeout() time.Duration {
	base := waitEpochTimeout()
	interval := readPositiveDurationEnv("PRUNER_INTERVAL", time.Minute)
	return base + 4*interval
}

func scanSidecarStorage(
	sidecarIndex int,
	targetEpoch symbiotic.Epoch,
	requestID common.Hash,
) ([]remainingEntity, error) {
	if strings.EqualFold(os.Getenv("STORAGE_TYPE"), "badger") {
		return scanBadgerStorage(sidecarIndex, targetEpoch, requestID)
	}
	return scanBboltStorage(sidecarIndex, targetEpoch, requestID)
}

func stopSidecarForStorageScan(t *testing.T, sidecarIndex int) {
	t.Helper()

	containerName := fmt.Sprintf("symbiotic-relay-%d", sidecarIndex+1)
	cmd := exec.CommandContext(t.Context(), "docker", "stop", "--time", "1", containerName)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "failed to stop %s: %s", containerName, string(output))
}

func badgerOfflineScanDelay() time.Duration {
	return 4 * readPositiveDurationEnv("PRUNER_INTERVAL", time.Minute)
}

func scanBboltStorage(
	sidecarIndex int,
	targetEpoch symbiotic.Epoch,
	requestID common.Hash,
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
