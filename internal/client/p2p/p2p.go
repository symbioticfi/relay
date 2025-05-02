package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/samber/lo"

	"middleware-offchain/internal/entity"
	log2 "middleware-offchain/pkg/log"
)

// Configuration
const (
	protocolID = "/p2p/messaging/1.0.0"
)

// Service handles peer-to-peer communication and signature aggregation
type Service struct {
	ctx            context.Context
	host           host.Host
	peersMutex     sync.RWMutex
	peers          map[peer.ID]struct{}
	messageHandler func(msg entity.P2PMessage) error
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, listenAddrs ...string) (*Service, error) {
	slog.InfoContext(ctx, "Listening on", "addrs", listenAddrs)

	h, err := libp2p.New(libp2p.ListenAddrStrings(listenAddrs...))
	if err != nil {
		return nil, errors.Errorf("failed to create libp2p host: %w", err)
	}

	service := &Service{
		ctx:   log2.WithAttrs(ctx, slog.String("component", "p2p")),
		host:  h,
		peers: make(map[peer.ID]struct{}),
	}
	h.SetStreamHandler(protocolID, service.HandleStream)
	h.Network().Notify(service)
	return service, nil
}

func (s *Service) HandleStream(stream network.Stream) {
	if err := s.handleStream(stream); err != nil {
		slog.ErrorContext(s.ctx, "Failed to handle stream", "error", err)
	}
}

func (s *Service) SetMessageHandler(mh func(msg entity.P2PMessage) error) {
	s.messageHandler = mh // todo ilya check if nil
}

func (s *Service) handleStream(stream network.Stream) error {
	defer stream.Close()

	data := make([]byte, 1024)
	n, err := stream.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}
	var message entity.P2PMessage
	if err := json.Unmarshal(data[:n], &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if err := s.messageHandler(message); err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}

	return nil
}

func (s *Service) AddPeer(pi peer.AddrInfo) error {
	if pi.ID == s.host.ID() {
		return nil
	}

	slog.InfoContext(s.ctx, "Trying to add peer", "peer", pi.ID)

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
	defer cancel()

	if err := s.host.Connect(ctx, pi); err != nil {
		return errors.Errorf("failed to connect to peer %s: %w", pi.ID.ShortString(), err)
	}

	s.peersMutex.Lock()
	if _, found := s.peers[pi.ID]; !found {
		s.peers[pi.ID] = struct{}{}
	}
	s.peersMutex.Unlock()

	slog.InfoContext(ctx, "Connected to peer", "peer", pi.ID, "totalPeers", len(s.peers))

	return nil
}

// Broadcast sends a message to all connected peers
func (s *Service) Broadcast(msg entity.P2PMessage) error {
	s.peersMutex.RLock()
	peers := lo.Keys(s.peers)
	s.peersMutex.RUnlock()

	for _, peerID := range peers {
		if err := s.sendToPeer(peerID, msg); err != nil {
			return err // todo ilya send parallel + join errors + timeout
		}
	}

	return nil
}

// sendToPeer sends a message to a specific peer
func (s *Service) sendToPeer(peerID peer.ID, msg entity.P2PMessage) error {
	// Open a stream to the peer
	stream, err := s.host.NewStream(s.ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Marshal and send the message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	return nil
}

// Close gracefully stops the service
func (s *Service) Close() error {
	if err := s.host.Close(); err != nil {
		return fmt.Errorf("failed to close host: %w", err)
	}

	return nil
}

func (s *Service) HostID() peer.ID { // todo ilya do we need this?
	return s.host.ID()
}

func (s *Service) Listen(n network.Network, multiaddr multiaddr.Multiaddr) {
}

func (s *Service) ListenClose(n network.Network, multiaddr multiaddr.Multiaddr) {
}

func (s *Service) Connected(n network.Network, conn network.Conn) {
}

func (s *Service) Disconnected(n network.Network, conn network.Conn) {
	s.peersMutex.Lock()
	delete(s.peers, conn.RemotePeer())
	s.peersMutex.Unlock()
	slog.InfoContext(s.ctx, "Disconnected from peer", "remotePeer", conn.RemotePeer(), "localPeer", conn.LocalPeer(), "totalPeers", len(s.peers))
}
