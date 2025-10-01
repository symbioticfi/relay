package p2p

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/go-errors/errors"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

// DiscoveryService implements peer discovery functionality allowing both DHT and mDNS.
type DiscoveryService struct {
	cfg  Config
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

// DiscoverySvc represents the peer discovery service interface.
type DiscoverySvc interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
	GetDiscoveryClient(ctx context.Context) *drouting.RoutingDiscovery
}

const ProtocolPrefix = "/symbiotic"

// NewDiscoveryService creates a new discovery service.
func NewDiscoveryService(cfg Config) (*DiscoveryService, error) {
	if cfg.Host == nil {
		return nil, errors.New("host cannot be nil")
	}

	service := &DiscoveryService{
		cfg:               cfg,
		host:              cfg.Host,
		dht:               nil,
		mdns:              nil,
		rdiscov:           nil,
		mu:                sync.RWMutex{},
		started:           false,
		ctx:               nil,
		cancel:            nil,
		lastAdvertisement: time.Time{},
		bootstrapPeers:    make([]peer.AddrInfo, 0),
	}

	return service, nil
}

func (s *DiscoveryService) initDHT(ctx context.Context) error {
	if s.cfg.Discovery.DHTMode == "disabled" {
		slog.InfoContext(ctx, "DHT disabled by configuration")
		return nil
	}
	mode := s.determineDHTMode(ctx)
	bootnodes, err := s.parseBootstrapPeers(ctx)
	if err != nil {
		return errors.Errorf("failed to parse bootstrap peers: %w", err)
	}
	s.bootstrapPeers = bootnodes
	kdht, err := dht.New(ctx, s.host,
		dht.Mode(mode),
		dht.ProtocolPrefix(ProtocolPrefix),
		dht.RoutingTableRefreshPeriod(s.cfg.Discovery.DHTRoutingTableRefreshInterval),
		dht.BootstrapPeers(s.bootstrapPeers...),
	)
	if err != nil {
		return errors.Errorf("failed to create DHT: %w", err)
	}
	s.dht = kdht
	s.rdiscov = drouting.NewRoutingDiscovery(kdht)

	if err := s.dht.Bootstrap(ctx); err != nil {
		return errors.Errorf("failed to bootstrap DHT: %w", err)
	}

	go s.maintainConnections(ctx)

	slog.InfoContext(ctx, "DHT initialized", "mode", mode, "bucket_size", 25, "concurrency", 20)
	return nil
}

func (s *DiscoveryService) determineDHTMode(ctx context.Context) dht.ModeOpt {
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

func (s *DiscoveryService) parseBootstrapPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	bootPeers := make([]peer.AddrInfo, 0, len(s.cfg.Discovery.BootstrapPeers))
	for _, peerAddr := range s.cfg.Discovery.BootstrapPeers {
		addrInfo, err := peer.AddrInfoFromString(peerAddr)
		if err != nil {
			return nil, errors.Errorf("invalid bootstrap peer address %s: %w", peerAddr, err)
		}
		if addrInfo.ID == s.host.ID() {
			slog.WarnContext(ctx, "Skipping self as bootstrap peer", "peer", addrInfo.ID)
			continue
		}
		bootPeers = append(bootPeers, *addrInfo)
	}
	return bootPeers, nil
}

func (s *DiscoveryService) initMDNS(ctx context.Context) error {
	if !s.cfg.Discovery.EnableMDNS {
		slog.InfoContext(ctx, "mDNS disabled by configuration")
		return nil
	}
	mdnsService := mdns.NewMdnsService(s.host, s.cfg.Discovery.MDNSServiceName, s)
	s.mdns = mdnsService
	if err := s.mdns.Start(); err != nil {
		return errors.Errorf("failed to start mDNS: %w", err)
	}

	slog.InfoContext(ctx, "mDNS initialized", "service-name", s.cfg.Discovery.MDNSServiceName)
	return nil
}

func (s *DiscoveryService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("discovery service already started")
	}

	wrappedCtx, cancel := context.WithCancel(ctx)
	s.ctx = wrappedCtx
	s.cancel = cancel

	if err := s.initDHT(wrappedCtx); err != nil {
		cancel()
		return errors.Errorf("failed to initialize DHT: %w", err)
	}

	if err := s.initMDNS(wrappedCtx); err != nil {
		cancel()
		return errors.Errorf("failed to initialize mDNS: %w", err)
	}

	s.started = true
	slog.InfoContext(wrappedCtx, "Discovery service started successfully")
	return nil
}

func (s *DiscoveryService) maintainConnections(ctx context.Context) {
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

func (s *DiscoveryService) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return nil
	}

	slog.InfoContext(ctx, "Stopping discovery service...")
	s.cancel()
	if s.dht != nil {
		if err := s.dht.Close(); err != nil {
			slog.ErrorContext(ctx, "Error closing DHT", "error", err)
		}
	}
	if s.mdns != nil {
		if err := s.mdns.Close(); err != nil {
			slog.ErrorContext(ctx, "Error closing mDNS", "error", err)
		}
	}
	s.started = false
	slog.InfoContext(ctx, "Discovery service stopped successfully")
	return nil
}

func (s *DiscoveryService) GetDiscoveryClient(ctx context.Context) *drouting.RoutingDiscovery {
	if s.rdiscov == nil {
		slog.WarnContext(ctx, "Routing discovery client is not initialized")
		return nil
	}
	slog.DebugContext(ctx, "Returning routing discovery client")
	return s.rdiscov
}

func (s *DiscoveryService) Advertise(ctx context.Context, topic string) error {
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
		return errors.Errorf("failed to advertise for topic %s: %w", topic, err)
	}
	slog.DebugContext(ctx, "Successfully advertised for topic", "topic", topic, "ttl", ttl)
	return nil
}

// HandlePeerFound processes a newly discovered mDNS peer and attempts to connect.
func (s *DiscoveryService) HandlePeerFound(peerInfo peer.AddrInfo) {
	// linter suggests to use separate context
	wrapperCtx := context.WithoutCancel(s.ctx)

	slog.DebugContext(wrapperCtx, "Discovered new mDNS peer", "peer", peerInfo.ID, "addresses", peerInfo.Addrs)

	// Attempt to connect to the discovered peer
	if err := s.connectToPeer(wrapperCtx, peerInfo); err != nil {
		slog.WarnContext(wrapperCtx, "Failed to connect to mDNS peer", "peer", peerInfo.ID, "error", err)
	}
}

func (s *DiscoveryService) connectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error {
	if s.host.Network().Connectedness(peerInfo.ID) == network.Connected {
		slog.DebugContext(ctx, "Already connected to peer", "peer", peerInfo.ID)
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.Discovery.ConnectionTimeout)
	defer cancel()
	return s.host.Connect(ctx, peerInfo)
}
