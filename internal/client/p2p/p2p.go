package p2p

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
	p2pEntity "middleware-offchain/internal/entity"
	"middleware-offchain/pkg/log"
)

// Configuration
const (
	signedHashProtocolID      protocol.ID = "/p2p/messaging/1.0.0/signedHash"
	aggregatedProofProtocolID protocol.ID = "/p2p/messaging/1.0.0/aggregatedProof"
)

type messageType string

const (
	messageTypeSignatureHash        messageType = "signature_hash_generated"
	messageTypeSignaturesAggregated messageType = "signatures_aggregated"
)

// Service handles peer-to-peer communication and signature aggregation
type Service struct {
	ctx                         context.Context
	host                        host.Host
	peersMutex                  sync.RWMutex
	peers                       map[peer.ID]struct{}
	signatureHashHandler        func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.SignatureMessage) error
	signaturesAggregatedHandler func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.AggregatedSignatureMessage) error
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, h host.Host) (*Service, error) {
	service := &Service{
		ctx:   log.WithAttrs(ctx, slog.String("component", "p2p")),
		host:  h,
		peers: make(map[peer.ID]struct{}),
		signatureHashHandler: func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.SignatureMessage) error {
			return nil
		},
		signaturesAggregatedHandler: func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.AggregatedSignatureMessage) error {
			return nil
		},
	}

	h.SetStreamHandler(signedHashProtocolID, handleStreamWrapper(ctx, service.handleStreamSignedHash))
	h.SetStreamHandler(aggregatedProofProtocolID, handleStreamWrapper(ctx, service.handleStreamAggregatedProof))

	h.Network().Notify(service)

	return service, nil
}

func (s *Service) SetSignatureHashMessageHandler(mh func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.SignatureMessage) error) {
	s.signatureHashHandler = mh // todo ilya check if nil + mutex
}

func (s *Service) SetSignaturesAggregatedMessageHandler(mh func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.AggregatedSignatureMessage) error) {
	s.signaturesAggregatedHandler = mh // todo ilya check if nil + mutex
}

func (s *Service) AddPeer(pi peer.AddrInfo) error {
	if pi.ID == s.host.ID() {
		slog.InfoContext(s.ctx, "Skipping self-connection", "peer", pi.ID)
		return nil
	}

	slog.DebugContext(s.ctx, "Trying to add peer", "peer", pi.ID, "addrs", pi.Addrs)

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
	defer cancel()

	if err := s.host.Connect(ctx, pi); err != nil {
		slog.ErrorContext(s.ctx, "Failed to connect to peer", "peer", pi.ID, "error", err)
		return errors.Errorf("failed to connect to peer %s: %w", pi.ID.ShortString(), err)
	}

	s.peersMutex.Lock()
	if _, found := s.peers[pi.ID]; !found {
		s.peers[pi.ID] = struct{}{}
	}

	slog.InfoContext(ctx, "Connected to peer", "peer", pi.ID, "totalPeers", len(s.peers))

	s.peersMutex.Unlock()

	return nil
}

// p2pMessage is the basic unit of communication between peers
type p2pMessage struct {
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
}

// broadcast sends a message to all connected peers
func (s *Service) broadcast(ctx context.Context, typ messageType, data []byte) error {
	s.peersMutex.RLock()
	peers := lo.Keys(s.peers)
	s.peersMutex.RUnlock()

	msg := p2pMessage{
		Sender:    s.host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	for _, peerID := range peers {
		if err := s.sendToPeer(ctx, typ, peerID, msg); err != nil {
			return err // todo ilya send parallel + join errors + timeout
		}
	}

	return nil
}

// sendToPeer sends a message to a specific peer
func (s *Service) sendToPeer(ctx context.Context, typ messageType, peerID peer.ID, msg p2pMessage) error {
	protocolID, err := getProtocolIDByMessageType(typ)
	if err != nil {
		return errors.Errorf("failed to get protocol ID: %w", err)
	}

	stream, err := s.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return errors.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Marshal and send the message
	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Errorf("failed to marshal message: %w", err)
	}

	_, err = stream.Write(data)
	if err != nil {
		return errors.Errorf("failed to write to stream: %w", err)
	}

	return nil
}

func getProtocolIDByMessageType(messageType messageType) (protocol.ID, error) {
	switch messageType {
	case messageTypeSignatureHash:
		return signedHashProtocolID, nil
	case messageTypeSignaturesAggregated:
		return aggregatedProofProtocolID, nil
	default:
		return "", errors.Errorf("unknown message type: %s", messageType)
	}
}

// Close gracefully stops the service
func (s *Service) Close() error {
	if err := s.host.Close(); err != nil {
		return errors.Errorf("failed to close host: %w", err)
	}

	return nil
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

	slog.InfoContext(s.ctx, "Disconnected from peer", "remotePeer", conn.RemotePeer(), "localPeer", conn.LocalPeer(), "totalPeers", len(s.peers))

	s.peersMutex.Unlock()
}
