package badger

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func randomAddr(t *testing.T) symbiotic.CrossChainAddress {
	t.Helper()

	chainID, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	return symbiotic.CrossChainAddress{
		ChainId: chainID.Uint64(),
		Address: common.BytesToAddress(randomBytes(t, 20)),
	}
}

func randomBigInt(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	require.NoError(t, err)
	return n
}

func TestKeyRequestIDEpoch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		epoch     symbiotic.Epoch
		requestID common.Hash
	}{
		{
			name:      "epoch 0",
			epoch:     0,
			requestID: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		},
		{
			name:      "epoch 1",
			epoch:     1,
			requestID: common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		},
		{
			name:      "epoch with colon byte (58)",
			epoch:     58,
			requestID: common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		},
		{
			name:      "epoch with colon byte in middle (14848 = 0x3A00)",
			epoch:     14848,
			requestID: common.HexToHash("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		},
		{
			name:      "large epoch",
			epoch:     999999999,
			requestID: common.HexToHash("0xdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
		},
		{
			name:      "max uint64 epoch",
			epoch:     ^symbiotic.Epoch(0), // max uint64
			requestID: common.HexToHash("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := keyRequestIDEpoch(tt.epoch, tt.requestID)

			prefix := keyRequestIDEpochPrefix(tt.epoch)
			require.Greater(t, len(key), len(prefix), "key must be longer than prefix")
			require.Equal(t, prefix, key[:len(prefix)], "key must start with prefix")

			extractedRequestID, err := extractRequestIDFromEpochKey(key)
			require.NoError(t, err, "must extract requestID without error")
			require.Equal(t, tt.requestID, extractedRequestID, "extracted requestID must match original")
		})
	}
}

func TestExtractRequestIDFromEpochKey_InvalidKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  []byte
	}{
		{
			name: "empty key",
			key:  []byte{},
		},
		{
			name: "too short key - only prefix",
			key:  keyRequestIDEpochAll(),
		},
		{
			name: "too short key - prefix + partial epoch",
			key:  append(keyRequestIDEpochAll(), []byte{1, 2, 3}...),
		},
		{
			name: "too short key - prefix + epoch but no requestID",
			key:  keyRequestIDEpochPrefix(1),
		},
		{
			name: "prefix + epoch + partial requestID",
			key:  append(keyRequestIDEpochPrefix(1), []byte("0x1234")...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractRequestIDFromEpochKey(tt.key)
			require.Error(t, err, "must return error for invalid key")
		})
	}
}

func TestKeyRequestIDEpochPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		epoch symbiotic.Epoch
	}{
		{
			name:  "epoch 0",
			epoch: 0,
		},
		{
			name:  "epoch 1",
			epoch: 1,
		},
		{
			name:  "epoch 58 (contains colon byte)",
			epoch: 58,
		},
		{
			name:  "epoch 14848 (0x3A00 - contains colon byte)",
			epoch: 14848,
		},
		{
			name:  "large epoch",
			epoch: 999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := keyRequestIDEpochPrefix(tt.epoch)

			basePrefix := keyRequestIDEpochAll()
			require.Equal(t, len(prefix), len(basePrefix)+epochLen, "prefix must be base + epoch length")
			require.Equal(t, basePrefix, prefix[:len(basePrefix)], "prefix must start with base prefix")

			epochBytes := tt.epoch.Bytes()
			require.Equal(t, epochBytes, prefix[len(basePrefix):], "epoch bytes must be present after base prefix")
		})
	}
}

func TestKeyRequestIDEpochAll(t *testing.T) {
	t.Parallel()

	prefix := keyRequestIDEpochAll()
	require.Equal(t, "request_id_epoch", string(prefix), "base prefix must be 'request_id_epoch'")
	require.Len(t, prefix, 16, "base prefix must be 16 bytes")
}
