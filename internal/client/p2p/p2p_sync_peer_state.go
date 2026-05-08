package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	syncPeerFailureBaseCooldown = 15 * time.Second
	syncPeerFailureMaxCooldown  = 2 * time.Minute
)

type peerSyncFailure struct {
	consecutiveFailures int
	cooldownUntil       time.Time
}

func (s *Service) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *Service) getEligibleSyncPeers(peers []peer.ID) []peer.ID {
	if len(peers) == 0 {
		return nil
	}

	now := s.currentTime()
	eligible := make([]peer.ID, 0, len(peers))

	s.peerSyncMu.RLock()
	defer s.peerSyncMu.RUnlock()

	for _, peerID := range peers {
		state, ok := s.peerSyncState[peerID]
		if !ok || !state.cooldownUntil.After(now) {
			eligible = append(eligible, peerID)
		}
	}

	return eligible
}

func (s *Service) markPeerSyncFailure(peerID peer.ID) {
	s.peerSyncMu.Lock()
	defer s.peerSyncMu.Unlock()

	if s.peerSyncState == nil {
		s.peerSyncState = make(map[peer.ID]peerSyncFailure)
	}

	state := s.peerSyncState[peerID]
	state.consecutiveFailures++
	state.cooldownUntil = s.currentTime().Add(nextSyncPeerCooldown(state.consecutiveFailures))
	s.peerSyncState[peerID] = state
}

func (s *Service) markPeerSyncSuccess(peerID peer.ID) {
	s.peerSyncMu.Lock()
	defer s.peerSyncMu.Unlock()

	if s.peerSyncState == nil {
		return
	}

	delete(s.peerSyncState, peerID)
}

func nextSyncPeerCooldown(failures int) time.Duration {
	if failures <= 1 {
		return syncPeerFailureBaseCooldown
	}

	cooldown := syncPeerFailureBaseCooldown
	for i := 1; i < failures; i++ {
		cooldown *= 2
		if cooldown >= syncPeerFailureMaxCooldown {
			return syncPeerFailureMaxCooldown
		}
	}

	return cooldown
}
