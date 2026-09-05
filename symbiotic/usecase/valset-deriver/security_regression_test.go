package valsetDeriver

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type canonicalHeaderClient struct {
	evmClient

	committed bool
	hash      common.Hash
	err       error
}

func (c canonicalHeaderClient) IsValsetHeaderCommittedAt(context.Context, symbiotic.CrossChainAddress, symbiotic.Epoch, ...symbiotic.EVMOption) (bool, error) {
	return c.committed, c.err
}
func (c canonicalHeaderClient) GetHeaderHashAt(context.Context, symbiotic.CrossChainAddress, symbiotic.Epoch) (common.Hash, error) {
	return c.hash, c.err
}

func TestCommittedHeaderValidation(t *testing.T) {
	valset := symbiotic.ValidatorSet{Version: 1, Epoch: 2, QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(0))}
	header, err := valset.GetHeader()
	require.NoError(t, err)
	hash, err := header.Hash()
	require.NoError(t, err)
	for _, tc := range []struct {
		name      string
		client    canonicalHeaderClient
		wantError bool
	}{
		{"uncommitted", canonicalHeaderClient{}, false},
		{"matching", canonicalHeaderClient{committed: true, hash: hash}, false},
		{"mismatch", canonicalHeaderClient{committed: true, hash: common.Hash{1}}, true},
		{"RPC error", canonicalHeaderClient{err: errors.New("offline")}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := Deriver{evmClient: tc.client}
			err := d.validateCommittedHeader(t.Context(), valset, []symbiotic.CrossChainAddress{{ChainId: 1}})
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegressionAggregateVotingPowerOverflow(t *testing.T) {
	power := new(big.Int).Lsh(big.NewInt(1), 255)
	inputs := []dtoOperatorVotingPower{}
	for chain := uint64(1); chain <= 2; chain++ {
		inputs = append(inputs, dtoOperatorVotingPower{chainId: chain, votingPowers: []symbiotic.OperatorVotingPower{{Operator: common.HexToAddress("0x01"), Vaults: []symbiotic.VaultVotingPower{{Vault: common.HexToAddress("0x02"), VotingPower: symbiotic.ToVotingPower(new(big.Int).Set(power))}}}}})
	}
	validators := fillValidators(inputs, nil)
	t.Logf("eachProviderBits=%d combinedValidatorBits=%d", power.BitLen(), validators[0].VotingPower.BitLen())
	require.NotPanics(t, func() {
		_, _, err := GetSchedulerInfo(t.Context(), symbiotic.ValidatorSet{Version: 1, Validators: validators, QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(0))}, symbiotic.NetworkConfig{})
		require.Error(t, err)
	})
}
