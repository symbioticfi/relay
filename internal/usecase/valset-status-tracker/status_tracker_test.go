package valsetStatusTracker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type cursorRepository struct {
	repo

	saved         symbiotic.Epoch
	noSettlements bool
}

func (r *cursorRepository) GetFirstUncommittedValidatorSetEpoch(context.Context) (symbiotic.Epoch, error) {
	return 1, nil
}
func (r *cursorRepository) GetLatestValidatorSetEpoch(context.Context) (symbiotic.Epoch, error) {
	return 1, nil
}
func (r *cursorRepository) GetConfigByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	if epoch == 1 && r.noSettlements {
		return symbiotic.NetworkConfig{}, nil
	}
	return symbiotic.NetworkConfig{Settlements: []symbiotic.CrossChainAddress{{ChainId: 1}}}, nil
}
func (r *cursorRepository) GetValidatorSetByEpoch(_ context.Context, e symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	if e == 1 {
		if r.noSettlements {
			return symbiotic.ValidatorSet{Epoch: 1, Status: symbiotic.HeaderDerived}, nil
		}
		return symbiotic.ValidatorSet{Epoch: 1, Status: symbiotic.HeaderCommitted}, nil
	}
	return symbiotic.ValidatorSet{}, entity.ErrEntityNotFound
}
func (r *cursorRepository) SaveFirstUncommittedValidatorSetEpoch(_ context.Context, e symbiotic.Epoch) error {
	r.saved = e
	return nil
}

type committedEpochChain struct{ evmClient }

func (committedEpochChain) GetCurrentEpoch(context.Context) (symbiotic.Epoch, error) { return 3, nil }
func (committedEpochChain) GetLastCommittedHeaderEpoch(context.Context, symbiotic.CrossChainAddress, ...symbiotic.EVMOption) (symbiotic.Epoch, error) {
	return 3, nil
}

func TestStatusCursorOnlyAdvancesThroughLocalEpochs(t *testing.T) {
	for _, noSettlements := range []bool{false, true} {
		name := "committed"
		if noSettlements {
			name = "no settlement obligations"
		}
		t.Run(name, func(t *testing.T) {
			r := &cursorRepository{noSettlements: noSettlements}
			s := &Service{cfg: Config{Repo: r, EvmClient: committedEpochChain{}}}
			require.NoError(t, s.trackCommittedEpochs(t.Context()))
			require.Equal(t, symbiotic.Epoch(2), r.saved)
		})
	}
}
