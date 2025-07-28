package p2p

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

// discService implements peer discovery functionality allowing both DHT and mDNS.
type discService struct {
	cfg  *Config
	host host.Host

	dht     *dht.IpfsDHT
	mdns    mdns.Service
	rdiscov *drouting.RoutingDiscovery

	mu                sync.RWMutex
	started           bool
	ctx               context.Context
	cancel            context.CancelFunc
	lastAdvertisement time.Time

	bootstrapPeers []peer.AddrInfo
}

// DiscoveryService represents the peer discovery service interface.
type DiscoveryService interface {
	Start() error
	Close() error
	GetDiscoveryClient(ctx context.Context) *drouting.RoutingDiscovery
}

const ProtocolPrefix = "/symbiotic"

// NewDiscoveryService creates a new discovery service.
func NewDiscoveryService(ctx context.Context, cfg *Config) (DiscoveryService, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if cfg.Discovery == nil {
		cfg.Discovery = DefaultDiscoveryConfig()
	}
	if cfg.Host == nil {
		return nil, errors.New("host cannot be nil")
	}

	discCtx, cancel := context.WithCancel(ctx)
	service := &discService{
		cfg:            cfg,
		host:           cfg.Host,
		ctx:            discCtx,
		cancel:         cancel,
		bootstrapPeers: make([]peer.AddrInfo, 0),
	}

	if err := service.initDHT(discCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize DHT: %w", err)
	}
	service.initMDNS(discCtx)

	if service.dht != nil {
		service.rdiscov = drouting.NewRoutingDiscovery(service.dht)
	}

	slog.InfoContext(discCtx, "Discovery service created successfully")
	return service, nil
}

func (s *discService) initDHT(ctx context.Context) error {
	if s.cfg.Discovery.DHTMode == "disabled" {
		slog.InfoContext(ctx, "DHT disabled by configuration")
		return nil
	}
	mode := s.determineDHTMode(ctx)
	bootnodes, err := s.parseBootstrapPeers(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}
	s.bootstrapPeers = bootnodes
	kdht, err := dht.New(ctx, s.host,
		dht.Mode(mode),
		dht.ProtocolPrefix(ProtocolPrefix),
		dht.RoutingTableRefreshPeriod(s.cfg.Discovery.DHTRoutingTableRefreshInterval),
		dht.BootstrapPeers(s.bootstrapPeers...),
	)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}
	s.dht = kdht
	slog.InfoContext(ctx, "DHT initialized", "mode", mode, "bucket_size", 25, "concurrency", 20)
	return nil
}

func (s *discService) determineDHTMode(ctx context.Context) dht.ModeOpt {
	switch s.cfg.Discovery.DHTMode {
	case "client":
		return dht.ModeClient
	case "server":
		return dht.ModeServer
	case "auto", "":
		return dht.ModeAuto
	default:
		slog.WarnContext(ctx, "Invalid DHT mode, defaulting to auto", "mode", s.cfg.Discovery.DHTMode)
		return dht.ModeAuto
	}
}

func (s *discService) parseBootstrapPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	bootPeers := make([]peer.AddrInfo, 0, len(s.cfg.Discovery.BootstrapPeers))
	for _, peerAddr := range s.cfg.Discovery.BootstrapPeers {
		addrInfo, err := peer.AddrInfoFromString(peerAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap peer address %s: %w", peerAddr, err)
		}
		if addrInfo.ID == s.host.ID() {
			slog.WarnContext(ctx, "Skipping self as bootstrap peer", "peer", addrInfo.ID)
			continue
		}
		bootPeers = append(bootPeers, *addrInfo)
	}
	return bootPeers, nil
}

func (s *discService) initMDNS(ctx context.Context) {
	if !s.cfg.Discovery.EnableMDNS {
		slog.InfoContext(ctx, "mDNS disabled by configuration")
	}
	mdnsService := mdns.NewMdnsService(s.host, s.cfg.Discovery.MDNSServiceName, s)
	s.mdns = mdnsService
	slog.InfoContext(ctx, "mDNS initialized", "service-name", s.cfg.Discovery.MDNSServiceName)
}

func (s *discService) Start() error {
	// linter suggests to use separate context
	wrapperCtx := context.WithoutCancel(s.ctx)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("discovery service already started")
	}
	slog.InfoContext(wrapperCtx, "Starting discovery service...")

	if s.dht != nil {
		if err := s.dht.Bootstrap(wrapperCtx); err != nil {
			return fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
		slog.InfoContext(wrapperCtx, "DHT bootstrap initiated")
		go s.maintainConnections(wrapperCtx)
	}
	if s.mdns != nil {
		if err := s.mdns.Start(); err != nil {
			return fmt.Errorf("failed to start mDNS: %w", err)
		}
		slog.InfoContext(wrapperCtx, "mDNS discovery started")
	}
	s.started = true
	slog.InfoContext(wrapperCtx, "Discovery service started successfully")
	return nil
}

func (s *discService) maintainConnections(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Discovery.DHTPeerDiscoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			for _, peerInfo := range s.bootstrapPeers {
				if err := s.connectToPeer(ctx, peerInfo); err != nil {
					slog.DebugContext(ctx, "Failed to connect to bootstrap peer", "peer", peerInfo.ID, "error", err)
				} else {
					slog.DebugContext(ctx, "Connected to bootstrap peer", "peer", peerInfo.ID)
				}
			}
			s.mu.RUnlock()

			if s.dht != nil && s.cfg.Discovery.AdvertiseServiceName != "" {
				if s.lastAdvertisement.IsZero() || time.Since(s.lastAdvertisement) >= s.cfg.Discovery.AdvertiseInterval {
					if err := s.Advertise(ctx, s.cfg.Discovery.AdvertiseServiceName); err != nil {
						slog.ErrorContext(ctx, "Failed to advertise in DHT", "error", err)
					} else {
						s.lastAdvertisement = time.Now()
						slog.DebugContext(ctx, "Successfully advertised in DHT", "peer", s.host.ID())
					}
				}

				findCtx, findCancel := context.WithCancel(ctx)
				peers, err := s.rdiscov.FindPeers(findCtx, s.cfg.Discovery.AdvertiseServiceName)
				if err != nil {
					findCancel()
					slog.ErrorContext(ctx, "Failed to find closest peers", "error", err)
					continue
				}
				count := s.cfg.Discovery.MaxDHTReconnectPeerCount
				for peerInfo := range peers {
					if s.host.ID() == peerInfo.ID {
						continue
					}
					if err := s.connectToPeer(ctx, peerInfo); err != nil {
						slog.WarnContext(ctx, "Failed to connect to DHT peer", "peer", peerInfo.ID, "error", err)
					} else {
						slog.DebugContext(ctx, "Connected to DHT peer", "peer", peerInfo.ID)
						count--
					}
					if count <= 0 {
						slog.DebugContext(ctx, "Reached connection limit for DHT peers", "limit", count)
						break
					}
				}
				findCancel()
			}
		}
	}
}

func (s *discService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return nil
	}

	// linter suggests to use separate context
	wrapperCtx := context.WithoutCancel(s.ctx)

	slog.InfoContext(wrapperCtx, "Stopping discovery service...")
	s.cancel()
	if s.dht != nil {
		if err := s.dht.Close(); err != nil {
			slog.ErrorContext(wrapperCtx, "Error closing DHT", "error", err)
		}
	}
	if s.mdns != nil {
		if err := s.mdns.Close(); err != nil {
			slog.ErrorContext(wrapperCtx, "Error closing mDNS", "error", err)
		}
	}
	s.started = false
	slog.InfoContext(wrapperCtx, "Discovery service stopped successfully")
	return nil
}

func (s *discService) GetDiscoveryClient(ctx context.Context) *drouting.RoutingDiscovery {
	if s.rdiscov == nil {
		slog.WarnContext(ctx, "Routing discovery client is not initialized")
		return nil
	}
	slog.DebugContext(ctx, "Returning routing discovery client")
	return s.rdiscov
}

func (s *discService) Advertise(ctx context.Context, topic string) error {
	if s.rdiscov == nil {
		return errors.New("routing discovery not available")
	}
	slog.DebugContext(ctx, "Advertising for topic", "topic", topic)
	ttl := s.cfg.Discovery.AdvertiseTTL
	if ttl == 0 {
		ttl = 12 * time.Hour
	}
	_, err := s.rdiscov.Advertise(ctx, topic, discovery.TTL(ttl))
	if err != nil {
		return fmt.Errorf("failed to advertise for topic %s: %w", topic, err)
	}
	slog.DebugContext(ctx, "Successfully advertised for topic", "topic", topic, "ttl", ttl)
	return nil
}

// HandlePeerFound processes a newly discovered mDNS peer and attempts to connect.
func (s *discService) HandlePeerFound(peerInfo peer.AddrInfo) {
	// linter suggests to use separate context
	wrapperCtx := context.WithoutCancel(s.ctx)

	slog.DebugContext(wrapperCtx, "Discovered new mDNS peer", "peer", peerInfo.ID, "addresses", peerInfo.Addrs)

	// Attempt to connect to the discovered peer
	if err := s.connectToPeer(wrapperCtx, peerInfo); err != nil {
		slog.WarnContext(wrapperCtx, "Failed to connect to mDNS peer", "peer", peerInfo.ID, "error", err)
	}
}

func (s *discService) connectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error {
	if s.host.Network().Connectedness(peerInfo.ID) == network.Connected {
		slog.DebugContext(ctx, "Already connected to peer", "peer", peerInfo.ID)
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.Discovery.ConnectionTimeout)
	defer cancel()
	return s.host.Connect(ctx, peerInfo)
}
