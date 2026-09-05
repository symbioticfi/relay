package valset_listener

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type regressionFailedChain struct{ evmClient }

func (regressionFailedChain) GetLastCommittedHeaderEpoch(_ context.Context, a symbiotic.CrossChainAddress, _ ...symbiotic.EVMOption) (symbiotic.Epoch, error) {
	if a.ChainId == 1 {
		return 0, errors.New("RPC unavailable")
	}
	return 50, nil
}
func TestCatchupDoesNotAdvancePastFailedSettlement(t *testing.T) {
	s := &Service{cfg: Config{EvmClient: regressionFailedChain{}}}
	got := s.detectLastCommittedEpochFromChain(t.Context(), symbiotic.NetworkConfig{Settlements: []symbiotic.CrossChainAddress{{ChainId: 1}, {ChainId: 2}}})
	t.Logf("chain1=unknown chain2=50 returnedMinimum=%d", got)
	require.Equal(t, symbiotic.Epoch(0), got)
}
