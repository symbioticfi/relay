package bbolt

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
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

func setupTestRepository(t *testing.T) *Repository {
	t.Helper()
	repo, err := New(Config{Dir: t.TempDir(), Metrics: repoutil.DoNothingMetrics{}})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, repo.Close())
	})
	return repo
}

func randomValidatorSet(t *testing.T, epoch symbiotic.Epoch) symbiotic.ValidatorSet {
	t.Helper()
	return symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(1000)),
		Validators: []symbiotic.Validator{
			{
				Operator:    common.BytesToAddress(randomBytes(t, 20)),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(15),
						Payload: randomBytes(t, 32),
					},
				},
				Vaults: []symbiotic.ValidatorVault{
					{
						ChainID:     1,
						Vault:       common.BytesToAddress(randomBytes(t, 20)),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					},
				},
			},
		},
		Status:            symbiotic.HeaderCommitted,
		AggregatorIndices: []uint32{},
		CommitterIndices:  []uint32{},
	}
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

func randomAggregationProof(t *testing.T) symbiotic.AggregationProof {
	t.Helper()
	return symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       10,
		Proof:       randomBytes(t, 32),
	}
}
