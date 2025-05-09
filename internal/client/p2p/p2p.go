package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-errors/errors"
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
	signatureHashProtocolID = "/p2p/messaging/1.0.0/signedHash"
)

// Service handles peer-to-peer communication and signature aggregation
type Service struct {
	ctx                         context.Context
	host                        host.Host
	peersMutex                  sync.RWMutex
	peers                       map[peer.ID]struct{}
	signatureHashMessageHandler func(ctx context.Context, msg entity.P2PSignatureHashMessage) error
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, h host.Host) (*Service, error) {
	service := &Service{
		ctx:   log2.WithAttrs(ctx, slog.String("component", "p2p")),
		host:  h,
		peers: make(map[peer.ID]struct{}),
		signatureHashMessageHandler: func(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
			return nil
		},
	}
	h.SetStreamHandler(signatureHashProtocolID, service.handleStream)
	h.Network().Notify(service)
	return service, nil
}

func (s *Service) handleStream(stream network.Stream) {
	if err := s.handleStreamInternal(stream); err != nil {
		slog.ErrorContext(s.ctx, "Failed to handle stream", "error", err)
	}
}

func (s *Service) SetSignatureHashMessageHandler(mh func(ctx context.Context, msg entity.P2PSignatureHashMessage) error) {
	s.signatureHashMessageHandler = mh // todo ilya check if nil + mutex
}

func (s *Service) handleStreamInternal(stream network.Stream) error {
	defer stream.Close()

	data := make([]byte, 1024*1024) // 1MB buffer
	n, err := stream.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %w", err)
	}

	var message p2pMessage
	if err := json.Unmarshal(data[:n], &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	switch message.Type {
	case entity.P2PMessageTypeSignatureHash:
		var signatureGenerated signatureGeneratedDTO
		if err := json.Unmarshal(message.Data, &signatureGenerated); err != nil {
			return fmt.Errorf("failed to unmarshal signatureGenerated message: %w", err)
		}
		entityMessage := entity.P2PSignatureHashMessage{
			Message: entity.SignatureHashMessage{
				MessageHash: signatureGenerated.MessageHash,
				Signature:   signatureGenerated.Signature,
				PublicKeyG1: signatureGenerated.PublicKeyG1,
				PublicKeyG2: signatureGenerated.PublicKeyG2,
				KeyTag:      signatureGenerated.KeyTag,
			},
			Info: entity.SenderInfo{
				Type:      message.Type,
				Sender:    message.Sender,
				Timestamp: message.Timestamp,
			},
		}
		if err := s.signatureHashMessageHandler(s.ctx, entityMessage); err != nil {
			return fmt.Errorf("failed to handle message: %w", err)
		}
	default:
		return fmt.Errorf("unknown message type: %s", message.Type)
	}

	return nil
}

func (s *Service) AddPeer(pi peer.AddrInfo) error {
	if pi.ID == s.host.ID() {
		slog.InfoContext(s.ctx, "Skipping self-connection", "peer", pi.ID)
		return nil
	}

	slog.InfoContext(s.ctx, "Trying to add peer", "peer", pi.ID, "addrs", pi.Addrs)

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
	s.peersMutex.Unlock()

	slog.InfoContext(ctx, "Connected to peer", "peer", pi.ID, "totalPeers", len(s.peers))

	return nil
}

// p2pMessage is the basic unit of communication between peers
type p2pMessage struct {
	Type      entity.P2PMessageType `json:"type"`
	Sender    string                `json:"sender"`
	Timestamp int64                 `json:"timestamp"`
	Data      []byte                `json:"data"`
}

// broadcast sends a message to all connected peers
func (s *Service) broadcast(ctx context.Context, typ entity.P2PMessageType, data []byte) error {
	s.peersMutex.RLock()
	peers := lo.Keys(s.peers)
	s.peersMutex.RUnlock()

	msg := p2pMessage{
		Type:      typ,
		Sender:    s.host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	for _, peerID := range peers {
		if err := s.sendToPeer(ctx, peerID, msg); err != nil {
			return err // todo ilya send parallel + join errors + timeout
		}
	}

	return nil
}

// sendToPeer sends a message to a specific peer
func (s *Service) sendToPeer(ctx context.Context, peerID peer.ID, msg p2pMessage) error {
	// Open a stream to the peer
	stream, err := s.host.NewStream(ctx, peerID, signatureHashProtocolID)
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
