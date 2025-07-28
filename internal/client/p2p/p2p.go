package p2p

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
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

type metrics interface {
	ObserveP2PMessageSent(messageType string)
	ObserveP2PPeerMessageSent(messageType, status string)
}

// DiscoveryConfig contains discovery protocol configuration
type DiscoveryConfig struct {
	// EnableMDNS specifies whether mDNS discovery is enabled.
	EnableMDNS bool `yaml:"enable_mdns"`
	// MDNSServiceName is the mDNS service name.
	MDNSServiceName string `yaml:"mdns_service_name"`

	// DHTMode specifies the DHT mode.
	DHTMode string `yaml:"dht_mode"`
	// BootstrapPeers is the list of bootstrap peers in multiaddr format.
	BootstrapPeers []string `yaml:"bootstrap_peers"`
	// AdvertiseTTL is the advertise time-to-live duration.
	AdvertiseTTL time.Duration `yaml:"advertise_ttl"`
	// AdvertiseServiceName is the advertise service name.
	AdvertiseServiceName string `yaml:"advertise_service_name"`
	// AdvertiseInterval is the interval between advertisements.
	AdvertiseInterval time.Duration `yaml:"advertise_period"`
	// ConnectionTimeout is the timeout for peer connections.
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	// MaxDHTReconnectPeerCount is the maximum number of DHT reconnect peers.
	MaxDHTReconnectPeerCount int `yaml:"max_dht_reconnect_peer_count"`
	// DHTPeerDiscoveryInterval is the interval for DHT peer discovery. Should be smaller than AdvertiseInterval.
	DHTPeerDiscoveryInterval time.Duration `yaml:"dht_peer_discovery_interval"`
	// DHTRoutingTableRefreshInterval is the interval for DHT routing table refresh. Should be greater than DHTPeerDiscoveryInterval.
	DHTRoutingTableRefreshInterval time.Duration `yaml:"dht_routing_table_refresh_interval"`
}

func DefaultDiscoveryConfig() *DiscoveryConfig {
	return &DiscoveryConfig{
		EnableMDNS:      false,
		MDNSServiceName: "symbiotic-mdns",

		DHTMode:                        "server",
		AdvertiseTTL:                   3 * time.Hour, // max allowed value in kdht package
		AdvertiseServiceName:           "symbiotic-advertise",
		AdvertiseInterval:              time.Hour,
		ConnectionTimeout:              5 * time.Second,
		MaxDHTReconnectPeerCount:       20,
		DHTPeerDiscoveryInterval:       5 * time.Minute,
		DHTRoutingTableRefreshInterval: 10 * time.Minute, // same as kdht package default
	}
}

type Config struct {
	Host        host.Host        `validate:"required"`
	SendTimeout time.Duration    `validate:"required,gt=0"`
	Metrics     metrics          `validate:"required"`
	Discovery   *DiscoveryConfig `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid P2P config: %w", err)
	}

	return nil
}

// Service handles peer-to-peer communication and signature aggregation
type Service struct {
	ctx                         context.Context
	host                        host.Host
	peersMutex                  sync.RWMutex
	peers                       map[peer.ID]struct{}
	signatureHashHandler        *signals.Signal[p2pEntity.P2PMessage[entity.SignatureMessage]]
	signaturesAggregatedHandler *signals.Signal[p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]]
	sendTimeout                 time.Duration
	metrics                     metrics
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	h := cfg.Host
	service := &Service{
		ctx:                         log.WithAttrs(ctx, slog.String("component", "p2p")),
		host:                        h,
		peers:                       make(map[peer.ID]struct{}),
		signatureHashHandler:        signals.New[p2pEntity.P2PMessage[entity.SignatureMessage]](),
		signaturesAggregatedHandler: signals.New[p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]](),
		sendTimeout:                 cfg.SendTimeout,
		metrics:                     cfg.Metrics,
	}

	h.SetStreamHandler(signedHashProtocolID, handleStreamWrapper(ctx, service.handleStreamSignedHash))
	h.SetStreamHandler(aggregatedProofProtocolID, handleStreamWrapper(ctx, service.handleStreamAggregatedProof))

	h.Network().Notify(service)

	return service, nil
}

func (s *Service) AddSignatureMessageListener(mh func(ctx context.Context, msg p2pEntity.P2PMessage[entity.SignatureMessage]) error, key string) {
	s.signatureHashHandler.AddListener(mh, key)
}

func (s *Service) AddSignaturesAggregatedMessageListener(mh func(ctx context.Context, msg p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]) error, key string) {
	s.signaturesAggregatedHandler.AddListener(mh, key)
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
	peers := s.host.Peerstore().Peers()
	s.peersMutex.RUnlock()

	msg := p2pMessage{
		Sender:    s.host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	wg := sync.WaitGroup{}
	errs := make([]error, len(peers))
	tmCtx, cancel := context.WithTimeout(ctx, s.sendTimeout)
	defer cancel()

	for i, peerID := range peers {
		if peerID == s.host.ID() {
			continue // Skip self
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if err := s.sendToPeer(tmCtx, typ, peerID, msg); err != nil {
				errs[i] = err
				s.metrics.ObserveP2PPeerMessageSent(string(typ), "error")
				return
			}

			s.metrics.ObserveP2PPeerMessageSent(string(typ), "ok")
		}(i)
	}

	wg.Wait()

	s.metrics.ObserveP2PMessageSent(string(typ))

	return errors.Join(errs...)
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

	if deadline, ok := ctx.Deadline(); ok {
		if err := stream.SetWriteDeadline(deadline); err != nil {
			return errors.Errorf("failed to set stream deadline: %w", err)
		}
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
