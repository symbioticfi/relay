package badger

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestBadgerRepository_UpdateSignatureStat(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	reqHash := common.BytesToHash(randomBytes(t, 32))
	now := time.Now().UTC()

	// Test creating new signature stat
	stat, err := repo.UpdateSignatureStat(t.Context(), reqHash, entity.SignatureStatStageAggCompleted, now)
	require.NoError(t, err)
	require.Equal(t, reqHash, stat.ReqHash)
	require.Equal(t, now.Truncate(time.Microsecond), stat.StatMap[entity.SignatureStatStageAggCompleted].Truncate(time.Microsecond))

	// Test updating existing signature stat with new stage
	later := now.Add(time.Hour)
	stat, err = repo.UpdateSignatureStat(t.Context(), reqHash, entity.SignatureStatStageSignCompleted, later)
	require.NoError(t, err)
	require.Equal(t, reqHash, stat.ReqHash)
	require.Len(t, stat.StatMap, 2)
	require.Equal(t, now.Truncate(time.Microsecond), stat.StatMap[entity.SignatureStatStageSignRequestReceived].Truncate(time.Microsecond))
	require.Equal(t, later.Truncate(time.Microsecond), stat.StatMap[entity.SignatureStatStageSignCompleted].Truncate(time.Microsecond))

	// Test updating existing stage with new timestamp
	evenLater := later.Add(time.Hour)
	stat, err = repo.UpdateSignatureStat(t.Context(), reqHash, entity.SignatureStatStageSignCompleted, evenLater)
	require.NoError(t, err)
	require.Equal(t, reqHash, stat.ReqHash)
	require.Len(t, stat.StatMap, 2)
	require.Equal(t, now.Truncate(time.Microsecond), stat.StatMap[entity.SignatureStatStageSignRequestReceived].Truncate(time.Microsecond))
	require.Equal(t, evenLater.Truncate(time.Microsecond), stat.StatMap[entity.SignatureStatStageSignCompleted].Truncate(time.Microsecond))
}
