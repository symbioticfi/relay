package badger

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"middleware-offchain/core/entity"
)

func TestBadgerRepository(t *testing.T) {
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	config := randomNetworkConfig(t)

	err = repo.SaveConfig(t.Context(), config, 1)
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
		Replicas:                []entity.CrossChainAddress{randomAddr(t)},
		VerificationType:        entity.VerificationTypeSimple,
		MaxVotingPower:          randomBigInt(t),
		MinInclusionVotingPower: randomBigInt(t),
		MaxValidatorsCount:      randomBigInt(t),
		RequiredKeyTags:         []entity.KeyTag{15},
	}
}

func randomAddr(t *testing.T) entity.CrossChainAddress {
	t.Helper()
	b := make([]byte, 20) // 20 bytes for Ethereum address
	_, err := rand.Read(b)
	require.NoError(t, err)

	chainID, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	return entity.CrossChainAddress{
		Address: common.BytesToAddress(b),
		ChainId: chainID.Uint64(),
	}
}

func randomBigInt(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	require.NoError(t, err)
	return n
}
