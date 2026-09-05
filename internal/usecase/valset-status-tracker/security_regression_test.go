package valsetStatusTracker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type regressionGapRepo struct {
	repo

	saved         symbiotic.Epoch
	visits        []symbiotic.Epoch
	noSettlements bool
}

func (r *regressionGapRepo) GetFirstUncommittedValidatorSetEpoch(context.Context) (symbiotic.Epoch, error) {
	return 1, nil
}
func (r *regressionGapRepo) GetLatestValidatorSetEpoch(context.Context) (symbiotic.Epoch, error) {
	return 1, nil
}
func (r *regressionGapRepo) GetConfigByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	if epoch == 1 && r.noSettlements {
		return symbiotic.NetworkConfig{}, nil
	}
	return symbiotic.NetworkConfig{Settlements: []symbiotic.CrossChainAddress{{ChainId: 1}}}, nil
}
func (r *regressionGapRepo) GetValidatorSetByEpoch(_ context.Context, e symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	if e == 1 {
		if r.noSettlements {
			return symbiotic.ValidatorSet{Epoch: 1, Status: symbiotic.HeaderDerived}, nil
		}
		return symbiotic.ValidatorSet{Epoch: 1, Status: symbiotic.HeaderCommitted}, nil
	}
	r.visits = append(r.visits, e)
	return symbiotic.ValidatorSet{}, entity.ErrEntityNotFound
}
func (r *regressionGapRepo) SaveFirstUncommittedValidatorSetEpoch(_ context.Context, e symbiotic.Epoch) error {
	r.saved = e
	return nil
}

type regressionGapChain struct{ evmClient }

func (regressionGapChain) GetCurrentEpoch(context.Context) (symbiotic.Epoch, error) { return 3, nil }
func (regressionGapChain) GetLastCommittedHeaderEpoch(context.Context, symbiotic.CrossChainAddress, ...symbiotic.EVMOption) (symbiotic.Epoch, error) {
	return 3, nil
}
func TestStatusCursorDoesNotSkipMissingValsets(t *testing.T) {
	r := &regressionGapRepo{}
	s := &Service{cfg: Config{Repo: r, EvmClient: regressionGapChain{}}}
	require.NoError(t, s.trackCommittedEpochs(t.Context()))
	t.Logf("missingEpochs=%v savedFirstUncommitted=%d", r.visits, r.saved)
	require.Equal(t, symbiotic.Epoch(2), r.saved)
}

func TestStatusCursorAdvancesPastEpochWithoutSettlements(t *testing.T) {
	r := &regressionGapRepo{noSettlements: true}
	s := &Service{cfg: Config{Repo: r, EvmClient: regressionGapChain{}}}
	require.NoError(t, s.trackCommittedEpochs(t.Context()))
	require.Equal(t, symbiotic.Epoch(2), r.saved)
}
