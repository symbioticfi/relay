package valset_listener

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type committedEpochClient struct {
	evmClient

	epochs map[uint64]symbiotic.Epoch
}

func (c committedEpochClient) GetLastCommittedHeaderEpoch(
	_ context.Context,
	addr symbiotic.CrossChainAddress,
	_ ...symbiotic.EVMOption,
) (symbiotic.Epoch, error) {
	return c.epochs[addr.ChainId], nil
}

func TestDetectLastCommittedEpochFromChainKeepsEpochZeroAsMinimum(t *testing.T) {
	settlements := []symbiotic.CrossChainAddress{
		{ChainId: 1, Address: common.HexToAddress("0x1111111111111111111111111111111111111111")},
		{ChainId: 2, Address: common.HexToAddress("0x2222222222222222222222222222222222222222")},
	}
	service := &Service{cfg: Config{EvmClient: committedEpochClient{
		epochs: map[uint64]symbiotic.Epoch{1: 0, 2: 5},
	}}}

	got := service.detectLastCommittedEpochFromChain(t.Context(), symbiotic.NetworkConfig{Settlements: settlements})

	require.Equal(t, symbiotic.Epoch(0), got)
}
