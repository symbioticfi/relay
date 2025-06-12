package badger

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"middleware-offchain/core/entity"
)

func TestBadgerRepository_NetworkConfig(t *testing.T) {
	t.Parallel()
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

func TestBadgerRepository_SignatureRequest(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	req := randomSignatureRequest(t)

	err = repo.SaveSignatureRequest(t.Context(), req)
	require.NoError(t, err)

	loadedConfig, err := repo.GetSignatureRequest(t.Context(), req.Hash())
	require.NoError(t, err)
	require.Equal(t, req, loadedConfig)
}

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

func randomSignatureRequest(t *testing.T) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: randomBigInt(t).Uint64(),
		Message:       randomBytes(t, 32),
	}
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

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func randomAddr(t *testing.T) entity.CrossChainAddress {
	t.Helper()

	chainID, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	return entity.CrossChainAddress{
		Address: common.BytesToAddress(randomBytes(t, 20)),
		ChainId: chainID.Uint64(),
	}
}

func randomBigInt(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	require.NoError(t, err)
	return n
}
