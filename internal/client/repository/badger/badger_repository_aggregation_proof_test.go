package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"middleware-offchain/core/entity"
)

func TestBadgerRepository_AggregationProof(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	ap := randomAggregationProof(t)

	hash := common.BytesToHash(randomBytes(t, 32))

	err = repo.SaveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)

	loadedConfig, err := repo.GetAggregationProof(t.Context(), hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedConfig)
}

func randomAggregationProof(t *testing.T) entity.AggregationProof {
	t.Helper()

	return entity.AggregationProof{
		VerificationType: entity.VerificationTypeSimple,
		MessageHash:      randomBytes(t, 32),
		Proof:            randomBytes(t, 32),
	}
}
