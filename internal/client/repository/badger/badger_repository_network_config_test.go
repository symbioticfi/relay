package badger

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_NetworkConfig(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	config := randomNetworkConfig(t)

	err := repo.saveConfig(t.Context(), config, 1)
	require.NoError(t, err)

	loadedConfig, err := repo.GetConfigByEpoch(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, config, loadedConfig)

	err = repo.saveConfig(t.Context(), config, 1)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
}

func randomNetworkConfig(t *testing.T) symbiotic.NetworkConfig {
	t.Helper()
	return symbiotic.NetworkConfig{
		VotingPowerProviders:    []symbiotic.CrossChainAddress{randomAddr(t)},
		KeysProvider:            randomAddr(t),
		Settlements:             []symbiotic.CrossChainAddress{randomAddr(t)},
		VerificationType:        symbiotic.VerificationTypeBlsBn254Simple,
		MaxVotingPower:          symbiotic.ToVotingPower(randomBigInt(t)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(randomBigInt(t)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(randomBigInt(t)),
		RequiredKeyTags:         []symbiotic.KeyTag{15},
		RequiredHeaderKeyTag:    7,
		QuorumThresholds: []symbiotic.QuorumThreshold{{
			KeyTag:          3,
			QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(123456789)),
		}},
		NumCommitters:  3,
		NumAggregators: 5,
	}
}
