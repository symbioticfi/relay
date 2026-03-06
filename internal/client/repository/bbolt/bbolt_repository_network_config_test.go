package bbolt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
)

func TestRepository_NetworkConfig(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	config := randomNetworkConfig(t)

	err := repo.SaveConfig(t.Context(), config, 1)
	require.NoError(t, err)

	loadedConfig, err := repo.GetConfigByEpoch(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, config, loadedConfig)

	err = repo.SaveConfig(t.Context(), config, 1)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
}
