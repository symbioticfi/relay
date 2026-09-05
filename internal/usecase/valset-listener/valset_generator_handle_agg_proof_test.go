package valset_listener

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type committedEpochClient struct {
	evmClient

	epochs      map[uint64]symbiotic.Epoch
	failedChain uint64
}

func (c committedEpochClient) GetLastCommittedHeaderEpoch(_ context.Context, addr symbiotic.CrossChainAddress, _ ...symbiotic.EVMOption) (symbiotic.Epoch, error) {
	if addr.ChainId == c.failedChain {
		return 0, errors.New("RPC unavailable")
	}
	return c.epochs[addr.ChainId], nil
}

func TestDetectLastCommittedEpochFromChain(t *testing.T) {
	for _, tc := range []struct {
		name        string
		epochs      map[uint64]symbiotic.Epoch
		failedChain uint64
		want        symbiotic.Epoch
	}{
		{"zero is a valid minimum", map[uint64]symbiotic.Epoch{1: 0, 2: 5}, 0, 0},
		{"all chains available", map[uint64]symbiotic.Epoch{1: 3, 2: 5}, 0, 3},
		{"first chain unavailable", map[uint64]symbiotic.Epoch{1: 3, 2: 5}, 1, 0},
		{"second chain unavailable", map[uint64]symbiotic.Epoch{1: 3, 2: 5}, 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &Service{cfg: Config{EvmClient: committedEpochClient{epochs: tc.epochs, failedChain: tc.failedChain}}}
			got := service.detectLastCommittedEpochFromChain(t.Context(), symbiotic.NetworkConfig{
				Settlements: []symbiotic.CrossChainAddress{{ChainId: 1}, {ChainId: 2}},
			})
			require.Equal(t, tc.want, got)
		})
	}
}
