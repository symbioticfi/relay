package bbolt

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestSignatureRequestPageByteBudget(t *testing.T) {
	repo := setupTestRepository(t)
	for i := byte(1); i <= 6; i++ {
		req := symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: 1, Message: bytes.Repeat([]byte{i}, 1024*1024)}
		require.NoError(t, repo.SaveSignatureRequest(t.Context(), common.BytesToHash([]byte{i}), req))
	}
	var cursor []byte
	seen := map[common.Hash]bool{}
	for pages := 0; ; pages++ {
		require.Less(t, pages, 6)
		rows, next, err := repo.GetSignatureRequestsWithIDByEpoch(t.Context(), 1, 1000, cursor)
		require.NoError(t, err)
		require.NotEmpty(t, rows)
		size := 0
		for _, row := range rows {
			require.False(t, seen[row.RequestID])
			seen[row.RequestID] = true
			size += len(row.SignatureRequest.Message)
		}
		require.LessOrEqual(t, size, repoutil.MaxSignatureRequestPageBytes)
		if next == nil {
			break
		}
		require.NotEqual(t, cursor, next)
		cursor = next
	}
	require.Len(t, seen, 6)
}
