package p2p

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/entity"
)

func TestP2P(t *testing.T) {
	v := atomic.Value{}
	ctx := t.Context()

	h1, err := libp2p.New()
	require.NoError(t, err)

	h2, err := libp2p.New()
	require.NoError(t, err)

	p2p1, err := NewService(ctx, h1)
	require.NoError(t, err)

	p2p2, err := NewService(ctx, h2)
	require.NoError(t, err)
	p2p2.SetSignatureHashMessageHandler(func(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
		v.Store(msg)
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

	err = p2p1.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: []byte("hello hash"),
		Signature:   []byte("hello signature"),
		PublicKeyG1: []byte("hello public key g1"),
		PublicKeyG2: []byte("hello public key g2"),
	})
	require.NoError(t, err)

	require.Eventuallyf(t, func() bool {
		msg := v.Load().(entity.P2PSignatureHashMessage)
		return msg.Info.Timestamp != 0
	}, time.Second, time.Millisecond*100, "Message not received in time")
	slog.InfoContext(ctx, "gotMessageIn2", "msg", v.Load())

	require.NoError(t, h1.Close())
	require.NoError(t, h2.Close())
}

func TestP2PMany(t *testing.T) {
	ctx := t.Context()

	mainHost, err := libp2p.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mainHost.Close()
	})

	mainService, err := NewService(ctx, mainHost)
	require.NoError(t, err)

	const countPeers = 10
	gotMessages := make([]entity.P2PSignatureHashMessage, countPeers)
	mu := sync.Mutex{}

	for i := 0; i < countPeers; i++ {
		otherHost, err := libp2p.New()
		require.NoError(t, err)
		t.Cleanup(func() {
			otherHost.Close()
		})

		p2p2, err := NewService(ctx, otherHost)
		require.NoError(t, err)
		p2p2.SetSignatureHashMessageHandler(func(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
			mu.Lock()
			defer mu.Unlock()

			gotMessages[i] = msg
			return nil
		})

		err = mainService.AddPeer(peer.AddrInfo{
			ID:    otherHost.ID(),
			Addrs: otherHost.Addrs(),
		})
		require.NoError(t, err)
	}

	err = mainService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: []byte("hello hash"),
		Signature:   []byte("hello signature"),
		PublicKeyG1: []byte("hello public key g1"),
		PublicKeyG2: []byte("hello public key g2"),
	})
	require.NoError(t, err)

	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()

		for _, message := range gotMessages {
			if message.Info.Timestamp == 0 {
				return false
			}
		}
		return true
	}, time.Second, time.Millisecond*100, "Message not received in time")
}
