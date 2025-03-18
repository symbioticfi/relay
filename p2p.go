package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"offchain-middleware/bls"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// Configuration
const (
	protocolID            = "/p2p/messaging/1.0.0"
	mdnsServiceTag        = "p2p-messaging"
	mdnsDiscoveryInterval = time.Second * 10
	signatureExpiration   = time.Minute * 5
	minSignatures         = 3 // Minimum signatures required for aggregation
)

// Message types
const (
	TypeSignatureRequest = "signature"
)

// Message is the basic unit of communication between peers
type Message struct {
	Type      string          `json:"type"`
	Sender    string          `json:"sender"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// TextMessage is a simple text message payload
type TextMessage struct {
	Content string `json:"content"`
}

// SignatureMessage contains a peer's signature for a message
type SignatureMessage struct {
	Epoch       *big.Int `json:"epoch"`
	MessageHash string   `json:"message_hash"`
	Signature   string   `json:"signature"`
	PublicKey   string   `json:"public_key"`
}

// P2PService handles peer-to-peer communication and signature aggregation
type P2PService struct {
	ctx          context.Context
	host         host.Host
	peersMutex   sync.RWMutex
	peers        map[peer.ID]struct{}
	blsSecretKey *bls.SecretKey
	storage      *Storage
}

// NewP2PService creates a new P2P service with the given configuration
func NewP2PService(ctx context.Context, listenAddrs []multiaddr.Multiaddr, storage *Storage, privateKeyHex string) (*P2PService, error) {
	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Generate BLS key pair using gnark BN254 curve
	secretKey, _, err := bls.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate BLS key pair: %w", err)
	}

	service := &P2PService{
		ctx:          ctx,
		host:         h,
		peers:        make(map[peer.ID]struct{}),
		blsSecretKey: secretKey,
		storage:      storage,
	}

	// Set up protocol handler
	h.SetStreamHandler(protocol.ID(protocolID), service.handleStream)

	// Print node info
	addrs := h.Addrs()
	for _, addr := range addrs {
		log.Printf("Listening on: %s/p2p/%s", addr, h.ID().ShortString())
	}

	return service, nil
}

// Start begins the service operations
func (s *P2PService) Start() error {
	// Start mDNS discovery
	discovery := mdns.NewMdnsService(s.host, mdnsServiceTag, s)
	if err := discovery.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS discovery service: %w", err)
	}

	return nil
}

// HandlePeerFound is called when a peer is discovered via mDNS
func (s *P2PService) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == s.host.ID() {
		return // Skip self
	}

	log.Printf("Discovered peer: %s", pi.ID.ShortString())

	s.peersMutex.Lock()
	if _, found := s.peers[pi.ID]; !found {
		s.peers[pi.ID] = struct{}{}
	}
	s.peersMutex.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, time.Second*10)
	defer cancel()

	if err := s.host.Connect(ctx, pi); err != nil {
		log.Printf("Failed to connect to peer %s: %s", pi.ID.ShortString(), err)
		return
	}

	log.Printf("Connected to peer: %s", pi.ID.ShortString())
}

// handleStream processes incoming streams from peers
func (s *P2PService) handleStream(stream network.Stream) {
	// Create a buffer for reading from the stream
	buf := make([]byte, 65536)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("Error reading from stream: %s", err)
		stream.Reset()
		return
	}

	// Parse the message
	var msg Message
	if err := json.Unmarshal(buf[:n], &msg); err != nil {
		log.Printf("Error unmarshaling message: %s", err)
		stream.Reset()
		return
	}

	// Process the message based on its type
	switch msg.Type {
	case TypeSignatureRequest:
		s.handleSignature(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	stream.Close()
}

// handleSignatureRequest processes a signature request
func (s *P2PService) handleSignature(msg Message) {
	var req SignatureMessage
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		log.Printf("Error unmarshaling signature request: %s", err)
		return
	}

	log.Printf("Received signature request for message: %s", req.MessageHash)

	// Convert message hash from hex to bytes
	msgHashBytes, err := hex.DecodeString(req.MessageHash)
	if err != nil {
		log.Printf("Error decoding message hash: %s", err)
		return
	}

	// Sign the message with BLS
	signature, err := s.blsSecretKey.Sign(msgHashBytes)
	if err != nil {
		log.Printf("Error signing message: %s", err)
		return
	}

	// Get public key
	publicKey := s.blsSecretKey.Public()
	pubKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		log.Printf("Error marshaling public key: %s", err)
		return
	}

	// Create the signature response
	msg := SignatureMessage{
		Epoch:       req.Epoch,
		MessageHash: req.MessageHash,
		Signature:   hex.EncodeToString(signature.Serialize()),
		PublicKey:   hex.EncodeToString(pubKeyBytes),
	}

	s.storage.AddSignature(req.Epoch, req.MessageHash, msg)
}

// BroadcastSignature broadcasts a signature request to all peers
func (s *P2PService) BroadcastSignature(epoch *big.Int, msgHash []byte) error {
	// Sign the message ourselves
	signature, err := s.blsSecretKey.Sign(msgHash)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	pubKeyBytes, err := s.blsSecretKey.Public().MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Create signature request
	req := SignatureMessage{
		Epoch:       epoch,
		MessageHash: hex.EncodeToString(msgHash),
		Signature:   hex.EncodeToString(signature.Serialize()),
		PublicKey:   hex.EncodeToString(pubKeyBytes),
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal signature request: %w", err)
	}

	msg := Message{
		Type:      TypeSignatureRequest,
		Sender:    s.host.ID().ShortString(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	if err := s.broadcast(msg); err != nil {
		return fmt.Errorf("failed to broadcast signature request: %w", err)
	}

	return nil
}

// broadcast sends a message to all connected peers
func (s *P2PService) broadcast(msg Message) error {
	s.peersMutex.RLock()
	defer s.peersMutex.RUnlock()

	for peerID := range s.peers {
		if err := s.sendToPeer(peerID.ShortString(), msg); err != nil {
			log.Printf("Failed to send message to peer %s: %s", peerID.ShortString(), err)
		}
	}

	return nil
}

// sendToPeer sends a message to a specific peer
func (s *P2PService) sendToPeer(peerIDStr string, msg Message) error {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Open a stream to the peer
	stream, err := s.host.NewStream(s.ctx, peerID, protocol.ID(protocolID))
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

// Stop gracefully stops the service
func (s *P2PService) Stop() error {
	if err := s.host.Close(); err != nil {
		return fmt.Errorf("failed to close host: %w", err)
	}
	return nil
}
