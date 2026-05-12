package p2p

import (
	"math"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
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
	state.cooldownUntil = s.currentTime().Add(s.nextSyncPeerCooldown(state.consecutiveFailures))
	state.consecutiveFailures++
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

func (s *Service) nextSyncPeerCooldown(failureCount int) time.Duration {
	if failureCount < 0 {
		failureCount = 0
	}

	cfg := s.syncPeerBackoff
	cooldown := math.Min(float64(cfg.MinBackoff)*math.Pow(cfg.Base, float64(failureCount)), float64(cfg.MaxBackoff))

	return time.Duration(cooldown)
}
