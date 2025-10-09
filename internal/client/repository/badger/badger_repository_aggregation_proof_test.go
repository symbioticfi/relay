package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_AggregationProof(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	ap := randomAggregationProof(t)

	hash := common.BytesToHash(randomBytes(t, 32))

	err := repo.saveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)
	err = repo.saveAggregationProof(t.Context(), hash, ap)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

	loadedConfig, err := repo.GetAggregationProof(t.Context(), hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedConfig)
}

func randomAggregationProof(t *testing.T) entity.AggregationProof {
	t.Helper()

	return entity.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      entity.KeyTag(15),
		Epoch:       10,
		Proof:       randomBytes(t, 32),
	}
}
