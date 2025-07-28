package p2p

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"

	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
)

const (
	topicPrefix = "/relay/v1"

	topicSignatureReady = topicPrefix + "/signature/ready"
	topicAggProofReady  = topicPrefix + "/proof/ready"
)

type metrics interface {
	ObserveP2PMessageSent(messageType string)
	ObserveP2PPeerMessageSent(messageType, status string)
}

// DiscoveryConfig contains discovery protocol configuration
type DiscoveryConfig struct {
	// EnableMDNS specifies whether mDNS discovery is enabled.
	EnableMDNS bool
	// MDNSServiceName is the mDNS service name.
	MDNSServiceName string

	// DHTMode specifies the DHT mode.
	DHTMode string
	// BootstrapPeers is the list of bootstrap peers in multiaddr format.
	BootstrapPeers []string
	// AdvertiseTTL is the advertise time-to-live duration.
	AdvertiseTTL time.Duration
	// AdvertiseServiceName is the advertise service name.
	AdvertiseServiceName string
	// AdvertiseInterval is the interval between advertisements.
	AdvertiseInterval time.Duration
	// ConnectionTimeout is the timeout for peer connections.
	ConnectionTimeout time.Duration
	// MaxDHTReconnectPeerCount is the maximum number of DHT reconnect peers.
	MaxDHTReconnectPeerCount int
	// DHTPeerDiscoveryInterval is the interval for DHT peer discovery. Should be smaller than AdvertiseInterval.
	DHTPeerDiscoveryInterval time.Duration
	// DHTRoutingTableRefreshInterval is the interval for DHT routing table refresh. Should be greater than DHTPeerDiscoveryInterval.
	DHTRoutingTableRefreshInterval time.Duration
}

func DefaultDiscoveryConfig() DiscoveryConfig {
	return DiscoveryConfig{
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
	Host        host.Host       `validate:"required"`
	SendTimeout time.Duration   `validate:"required,gt=0"`
	Metrics     metrics         `validate:"required"`
	Discovery   DiscoveryConfig `validate:"required"`
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
	signatureReceivedHandler    *signals.Signal[p2pEntity.P2PMessage[entity.SignatureMessage]]
	signaturesAggregatedHandler *signals.Signal[p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]]
	sendTimeout                 time.Duration
	metrics                     metrics
	topicsMap                   map[string]*pubsub.Topic
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	h := cfg.Host

	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
	)
	if err != nil {
		return nil, errors.Errorf("failed to create GossipSub: %w", err)
	}

	signatureReadyTopic, err := ps.Join(topicSignatureReady)
	if err != nil {
		return nil, errors.Errorf("failed to join signature ready topic: %w", err)
	}
	signatureReadySub, err := signatureReadyTopic.Subscribe()
	if err != nil {
		return nil, errors.Errorf("failed to subscribe to signature ready topic: %w", err)
	}

	proofReadyTopic, err := ps.Join(topicAggProofReady)
	if err != nil {
		return nil, errors.Errorf("failed to join agg proof ready topic: %w", err)
	}
	proofReadySub, err := proofReadyTopic.Subscribe()
	if err != nil {
		return nil, errors.Errorf("failed to subscribe to agg proof ready topic: %w", err)
	}

	service := &Service{
		ctx:                         log.WithAttrs(ctx, slog.String("component", "p2p")),
		host:                        h,
		signatureReceivedHandler:    signals.New[p2pEntity.P2PMessage[entity.SignatureMessage]](),
		signaturesAggregatedHandler: signals.New[p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]](),
		sendTimeout:                 cfg.SendTimeout,
		metrics:                     cfg.Metrics,

		topicsMap: map[string]*pubsub.Topic{
			topicSignatureReady: signatureReadyTopic,
			topicAggProofReady:  proofReadyTopic,
		},
	}

	go service.listenForMessages(ctx, signatureReadySub, service.handleSignatureReadyMessage)
	go service.listenForMessages(ctx, proofReadySub, service.handleAggregatedProofReadyMessage)

	h.Network().Notify(service)

	return service, nil
}

func (s *Service) listenForMessages(ctx context.Context, sub *pubsub.Subscription, handler func(ctx context.Context, msg *pubsub.Message) error) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read from subscription", "error", err)
			return
		}
		slog.DebugContext(ctx, "Received message from p2p", "topic", msg.Topic, "from", msg.ReceivedFrom, "data", string(msg.Data))
		if err := handler(ctx, msg); err != nil {
			slog.ErrorContext(ctx, "Failed to handle message", "error", err, "message", msg)
			continue
		}
	}
}

func (s *Service) AddSignatureMessageListener(mh func(ctx context.Context, msg p2pEntity.P2PMessage[entity.SignatureMessage]) error, key string) {
	s.signatureReceivedHandler.AddListener(mh, key)
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

	s.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)

	slog.InfoContext(ctx, "Connected to peer", "peer", pi.ID, "totalPeers", len(s.host.Peerstore().Peers()))

	return nil
}

// p2pMessage is the basic unit of communication between peers
type p2pMessage struct {
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
}

// broadcast sends a message to all connected peers
func (s *Service) broadcast(ctx context.Context, topicName string, data []byte) error {
	topic, ok := s.topicsMap[topicName]
	if !ok {
		return errors.Errorf("topic %s not found", topicName)
	}

	msg := p2pMessage{
		Sender:    s.host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	// Marshal and send the message
	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Errorf("failed to marshal message: %w", err)
	}

	err = topic.Publish(ctx, data)
	if err != nil {
		s.metrics.ObserveP2PPeerMessageSent(topicName, "error")
		return errors.Errorf("failed to publish data to topic %s: %w", topic.String(), err)
	}

	slog.DebugContext(ctx, "Message published to topic", "topic", topicName, "data", string(data))
	s.metrics.ObserveP2PPeerMessageSent(topicName, "ok")

	return nil
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
	s.host.Peerstore().RemovePeer(conn.RemotePeer())

	slog.InfoContext(s.ctx, "Disconnected from peer", "remotePeer", conn.RemotePeer(), "localPeer", conn.LocalPeer(), "totalPeers", len(s.host.Peerstore().Peers()))
}
