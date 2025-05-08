package p2p

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

func TestP2P(t *testing.T) {
	var gotMessageIn2 entity.P2PMessage
	ctx := context.Background()

	h1, err := libp2p.New()
	require.NoError(t, err)

	h2, err := libp2p.New()
	require.NoError(t, err)

	p2p1, err := NewService(ctx, h1)
	require.NoError(t, err)

	p2p2, err := NewService(ctx, h2)
	require.NoError(t, err)
	p2p2.SetMessageHandler(func(msg entity.P2PMessage) error {
		gotMessageIn2 = msg
		return nil
	})

	err = p2p1.AddPeer(peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})
	require.NoError(t, err)

	err = p2p2.AddPeer(peer.AddrInfo{
		ID:    h1.ID(),
		Addrs: h1.Addrs(),
	})
	require.NoError(t, err)

	err = p2p1.Broadcast(entity.TypeValsetGenerated, &types.ValidatorSetHeader{
		Version:                33,
		ActiveAggregatedKeys:   nil,
		TotalActiveVotingPower: nil,
		ValidatorsSszMRoot:     [32]byte{},
		ExtraData:              []byte("hello"),
	})
	require.NoError(t, err)

	require.Eventuallyf(t, func() bool {
		return gotMessageIn2.Timestamp != 0
	}, time.Second, time.Millisecond*100, "Message not received in time")
	slog.InfoContext(ctx, "gotMessageIn2", "msg", gotMessageIn2)

	require.NoError(t, h1.Close())
	require.NoError(t, h2.Close())
}

func TestP2PMany(t *testing.T) {
	ctx := context.Background()

	mainHost, err := libp2p.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mainHost.Close()
	})

	mainService, err := NewService(ctx, mainHost)
	require.NoError(t, err)

	const countPeers = 10
	gotMessages := make([]entity.P2PMessage, countPeers)
	for i := 0; i < countPeers; i++ {
		otherHost, err := libp2p.New()
		require.NoError(t, err)
		t.Cleanup(func() {
			otherHost.Close()
		})

		p2p2, err := NewService(ctx, otherHost)
		require.NoError(t, err)
		p2p2.SetMessageHandler(func(msg entity.P2PMessage) error {
			gotMessages[i] = msg
			return nil
		})

		err = mainService.AddPeer(peer.AddrInfo{
			ID:    otherHost.ID(),
			Addrs: otherHost.Addrs(),
		})
		require.NoError(t, err)
	}

	err = mainService.Broadcast(entity.TypeValsetGenerated, &types.ValidatorSetHeader{
		Version:                33,
		ActiveAggregatedKeys:   nil,
		TotalActiveVotingPower: nil,
		ValidatorsSszMRoot:     [32]byte{},
		ExtraData:              []byte("hello"),
	})
	require.NoError(t, err)

	require.Eventuallyf(t, func() bool {
		for _, message := range gotMessages {
			if message.Timestamp == 0 {
				return false
			}
		}
		return true
	}, time.Second, time.Millisecond*100, "Message not received in time")
	slog.InfoContext(ctx, "gotMessageIn2", "msg", gotMessages)
}
