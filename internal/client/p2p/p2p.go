package p2p

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/symbioticfi/relay/core/entity"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
)

const (
	topicPrefix = "/relay/v1"

	topicSignatureReady = topicPrefix + "/signature/ready"
	topicAggProofReady  = topicPrefix + "/proof/ready"

	maxP2PMessageSize  = 1<<20 + 1024 // 1 MiB + 1 KiB for overhead
	maxRequestHashSize = 32
	maxPubKeySize      = 96
	maxSignatureSize   = 96
	maxMsgHashSize     = 64
	maxProofSize       = 1 << 20
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
	Host            host.Host `validate:"required"`
	SkipMessageSign bool
	Metrics         metrics         `validate:"required"`
	Discovery       DiscoveryConfig `validate:"required"`
	EventTracer     pubsub.EventTracer
	Handler         prototypes.SymbioticP2PServiceServer `validate:"required"`
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
	metrics                     metrics
	topicsMap                   map[string]*pubsub.Topic
	p2pGRPCHandler              prototypes.SymbioticP2PServiceServer
}

// NewService creates a new P2P service with the given configuration
func NewService(ctx context.Context, cfg Config, signalCfg signals.Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	ctx = log.WithComponent(ctx, "p2p")

	h := cfg.Host

	signPolicy := pubsub.StrictSign
	if cfg.SkipMessageSign {
		slog.WarnContext(ctx, "Message signing is disabled, this may lead to security issues")
		signPolicy = pubsub.StrictNoSign
	}
	opts := []pubsub.Option{
		pubsub.WithMessageSignaturePolicy(signPolicy),
		pubsub.WithMaxMessageSize(maxP2PMessageSize),
	}
	if cfg.EventTracer != nil {
		opts = append(opts, pubsub.WithEventTracer(cfg.EventTracer))
	}
	ps, err := pubsub.NewGossipSub(ctx, h, opts...)
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
		signatureReceivedHandler:    signals.New[p2pEntity.P2PMessage[entity.SignatureMessage]](signalCfg, "signatureReceive", nil),
		signaturesAggregatedHandler: signals.New[p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]](signalCfg, "signaturesAggregated", nil),
		metrics:                     cfg.Metrics,

		topicsMap: map[string]*pubsub.Topic{
			topicSignatureReady: signatureReadyTopic,
			topicAggProofReady:  proofReadyTopic,
		},
		p2pGRPCHandler: cfg.Handler,
	}

	go service.listenForMessages(ctx, signatureReadySub, signatureReadyTopic, service.handleSignatureReadyMessage)
	go service.listenForMessages(ctx, proofReadySub, proofReadyTopic, service.handleAggregatedProofReadyMessage)

	h.Network().Notify(service)

	return service, nil
}

func (s *Service) listenForMessages(ctx context.Context, sub *pubsub.Subscription, topic *pubsub.Topic, handler func(ctx context.Context, msg *pubsub.Message) error) {
	slog.DebugContext(ctx, "Starting message listener", "topic", topic.String())
	defer func() {
		if err := topic.Close(); err != nil && !errors.Is(err, context.Canceled) {
			slog.WarnContext(ctx, "Failed to close topic", "topic", topic.String(), "error", err)
		}
		sub.Cancel()
		slog.InfoContext(ctx, "Subscription and topic closed", "topic", topic.String())
	}()

	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.InfoContext(ctx, "Subscription context canceled, stopping listener", "topic", topic.String())
				return
			}
			slog.ErrorContext(ctx, "Failed to read from subscription", "error", err)
			return
		}

		slog.DebugContext(ctx, "Received message from p2p", "topic", msg.Topic, "from", msg.ReceivedFrom)
		if err := handler(ctx, msg); err != nil {
			slog.ErrorContext(ctx, "Failed to handle message", "error", err, "message", msg)
			continue
		}
	}
}

func (s *Service) StartSignatureMessageListener(mh func(ctx context.Context, msg p2pEntity.P2PMessage[entity.SignatureMessage]) error) error {
	if err := s.signatureReceivedHandler.SetHandler(mh); err != nil {
		return errors.Errorf("failed to set signature received message handler: %w", err)
	}
	return s.signatureReceivedHandler.StartWorkers(s.ctx)
}

func (s *Service) StartSignaturesAggregatedMessageListener(mh func(ctx context.Context, msg p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]) error) error {
	if err := s.signaturesAggregatedHandler.SetHandler(mh); err != nil {
		return errors.Errorf("failed to set signatures aggregated message handler: %w", err)
	}
	return s.signaturesAggregatedHandler.StartWorkers(s.ctx)
}

func (s *Service) addPeer(pi peer.AddrInfo) error {
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
	return nil
}

// broadcast sends a message to all connected peers
func (s *Service) broadcast(ctx context.Context, topicName string, data []byte) error {
	topic, ok := s.topicsMap[topicName]
	if !ok {
		return errors.Errorf("topic %s not found", topicName)
	}

	msg := prototypes.P2PMessage{
		Sender:    s.host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	// Marshal and send the message
	data, err := proto.Marshal(&msg)
	if err != nil {
		return errors.Errorf("failed to marshal message: %w", err)
	}

	err = topic.Publish(ctx, data)
	if err != nil {
		s.metrics.ObserveP2PPeerMessageSent(topicName, "error")
		return errors.Errorf("failed to publish data to topic %s: %w", topic.String(), err)
	}

	slog.DebugContext(ctx, "Message published to topic", "topic", topicName)
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
	slog.DebugContext(s.ctx, "Listening on multiaddr", "multiaddr", multiaddr.String())
}

func (s *Service) ListenClose(n network.Network, multiaddr multiaddr.Multiaddr) {
	slog.DebugContext(s.ctx, "Stopped listening on multiaddr", "multiaddr", multiaddr.String())
}

func (s *Service) Connected(n network.Network, conn network.Conn) {
	slog.DebugContext(s.ctx, "Connected to peer", "peer", conn.RemotePeer().String(), "totalPeers", len(s.host.Peerstore().Peers()))
}

func (s *Service) Disconnected(n network.Network, conn network.Conn) {
	slog.DebugContext(s.ctx, "Disconnected from peer", "remotePeer", conn.RemotePeer(), "localPeer", conn.LocalPeer(), "totalPeers", len(s.host.Peerstore().Peers()))
}

func (s *Service) ID() string {
	return s.host.ID().String()
}

func (s *Service) StartGRPCServer(ctx context.Context) error {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(),
		grpc.ChainStreamInterceptor(),
	)
	prototypes.RegisterSymbioticP2PServiceServer(grpcServer, s.p2pGRPCHandler)

	listener, err := gostream.Listen(s.host, grpcProtocolTag)
	if err != nil {
		return errors.Errorf("failed to create gostream listener: %w", err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)
	defer close(serverErr)
	go func() {
		slog.InfoContext(ctx, "Starting gRPC server for P2P sync", "protocol", grpcProtocolTag)
		if err := grpcServer.Serve(listener); err != nil {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Shutting down gRPC server gracefully")
		grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serverErr:
		return errors.Errorf("gRPC server error: %w", err)
	}
}
