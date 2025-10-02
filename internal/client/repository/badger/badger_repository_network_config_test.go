package badger

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestBadgerRepository_NetworkConfig(t *testing.T) {
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

func randomNetworkConfig(t *testing.T) entity.NetworkConfig {
	t.Helper()
	return entity.NetworkConfig{
		VotingPowerProviders:    []entity.CrossChainAddress{randomAddr(t)},
		KeysProvider:            randomAddr(t),
		Settlements:             []entity.CrossChainAddress{randomAddr(t)},
		VerificationType:        entity.VerificationTypeBlsBn254Simple,
		MaxVotingPower:          entity.ToVotingPower(randomBigInt(t)),
		MinInclusionVotingPower: entity.ToVotingPower(randomBigInt(t)),
		MaxValidatorsCount:      entity.ToVotingPower(randomBigInt(t)),
		RequiredKeyTags:         []entity.KeyTag{15},
		RequiredHeaderKeyTag:    7,
		QuorumThresholds: []entity.QuorumThreshold{{
			KeyTag:          3,
			QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(123456789)),
		}},
		NumCommitters:  3,
		NumAggregators: 5,
	}
}
