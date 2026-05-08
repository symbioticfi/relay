package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/require"
)

func TestSelectPeerForSync_SkipsCooledDownPeersWhenPossible(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	service, peerIDs, cleanup := newServiceWithConnectedPeers(t, 2, now)
	defer cleanup()

	service.peerSyncState[peerIDs[0]] = peerSyncFailure{
		consecutiveFailures: 1,
		cooldownUntil:       now.Add(time.Minute),
	}

	selectedPeer, err := service.selectPeerForSync()
	require.NoError(t, err)
	require.Equal(t, peerIDs[1], selectedPeer)
}

func TestSelectPeerForSync_FallsBackWhenAllPeersAreCooledDown(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	service, peerIDs, cleanup := newServiceWithConnectedPeers(t, 2, now)
	defer cleanup()

	for _, peerID := range peerIDs {
		service.peerSyncState[peerID] = peerSyncFailure{
			consecutiveFailures: 1,
			cooldownUntil:       now.Add(time.Minute),
		}
	}

	selectedPeer, err := service.selectPeerForSync()
	require.NoError(t, err)
	require.Contains(t, peerIDs, selectedPeer)
}

func TestMarkPeerSyncFailure_IncreasesCooldown(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	service := &Service{
		peerSyncState: make(map[peer.ID]peerSyncFailure),
		now: func() time.Time {
			return now
		},
	}

	peerID := peer.ID("peer-1")

	service.markPeerSyncFailure(peerID)

	state := service.peerSyncState[peerID]
	require.Equal(t, 1, state.consecutiveFailures)
	require.Equal(t, now.Add(syncPeerFailureBaseCooldown), state.cooldownUntil)

	now = now.Add(syncPeerFailureBaseCooldown + time.Second)

	service.markPeerSyncFailure(peerID)

	state = service.peerSyncState[peerID]
	require.Equal(t, 2, state.consecutiveFailures)
	require.Equal(t, now.Add(syncPeerFailureBaseCooldown*2), state.cooldownUntil)
}

func TestMarkPeerSyncSuccess_ClearsFailureState(t *testing.T) {
	service := &Service{
		peerSyncState: map[peer.ID]peerSyncFailure{
			peer.ID("peer-1"): {
				consecutiveFailures: 2,
				cooldownUntil:       time.Now().Add(time.Minute),
			},
		},
	}

	service.markPeerSyncSuccess(peer.ID("peer-1"))

	require.Empty(t, service.peerSyncState)
}

func newServiceWithConnectedPeers(t *testing.T, peerCount int, now time.Time) (*Service, []peer.ID, func()) {
	t.Helper()

	clientHost, err := libp2p.New()
	require.NoError(t, err)

	peerHosts := make([]host.Host, 0, peerCount)
	peerIDs := make([]peer.ID, 0, peerCount)

	for range peerCount {
		remoteHost, err := libp2p.New()
		require.NoError(t, err)

		clientHost.Peerstore().AddAddrs(remoteHost.ID(), remoteHost.Addrs(), peerstore.PermanentAddrTTL)
		remoteHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)

		err = clientHost.Connect(context.Background(), peer.AddrInfo{
			ID:    remoteHost.ID(),
			Addrs: remoteHost.Addrs(),
		})
		require.NoError(t, err)

		peerHosts = append(peerHosts, remoteHost)
		peerIDs = append(peerIDs, remoteHost.ID())
	}

	service := &Service{
		host:          clientHost,
		peerSyncState: make(map[peer.ID]peerSyncFailure),
		now: func() time.Time {
			return now
		},
	}

	cleanup := func() {
		for _, remoteHost := range peerHosts {
			require.NoError(t, remoteHost.Close())
		}
		require.NoError(t, clientHost.Close())
	}

	return service, peerIDs, cleanup
}
