package p2p

import (
	"context"
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
	Stop() error
	GetDiscoveryClient() *drouting.RoutingDiscovery
}

const ProtocolPrefix = "/symbiotic"

// NewDiscoveryService creates a new discovery service.
func NewDiscoveryService(ctx context.Context, cfg *Config) (DiscoveryService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.Discovery == nil {
		cfg.Discovery = DefaultDiscoveryConfig()
	}
	if cfg.Host == nil {
		return nil, fmt.Errorf("host cannot be nil")
	}

	discCtx, cancel := context.WithCancel(ctx)
	service := &discService{
		cfg:            cfg,
		host:           cfg.Host,
		ctx:            discCtx,
		cancel:         cancel,
		bootstrapPeers: make([]peer.AddrInfo, 0),
	}

	if err := service.initDHT(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize DHT: %w", err)
	}
	if err := service.initMDNS(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize mDNS: %w", err)
	}
	if service.dht != nil {
		service.rdiscov = drouting.NewRoutingDiscovery(service.dht)
	}

	slog.Info("Discovery service created successfully")
	return service, nil
}

func (s *discService) initDHT() error {
	if s.cfg.Discovery.DHTMode == "disabled" {
		slog.Info("DHT disabled by configuration")
		return nil
	}
	mode := s.determineDHTMode()
	bootnodes, err := s.parseBootstrapPeers()
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}
	s.bootstrapPeers = bootnodes
	kdht, err := dht.New(s.ctx, s.host,
		dht.Mode(mode),
		dht.ProtocolPrefix(ProtocolPrefix),
		dht.RoutingTableRefreshPeriod(s.cfg.Discovery.DHTRoutingTableRefreshInterval),
		dht.BootstrapPeers(s.bootstrapPeers...),
	)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}
	s.dht = kdht
	slog.Info("DHT initialized", "mode", mode, "bucket_size", 25, "concurrency", 20)
	return nil
}

func (s *discService) determineDHTMode() dht.ModeOpt {
	switch s.cfg.Discovery.DHTMode {
	case "client":
		return dht.ModeClient
	case "server":
		return dht.ModeServer
	case "auto", "":
		return dht.ModeAuto
	default:
		slog.Warn("Invalid DHT mode, defaulting to auto", "mode", s.cfg.Discovery.DHTMode)
		return dht.ModeAuto
	}
}

func (s *discService) parseBootstrapPeers() ([]peer.AddrInfo, error) {
	var bootPeers []peer.AddrInfo
	for _, peerAddr := range s.cfg.Discovery.BootstrapPeers {
		addrInfo, err := peer.AddrInfoFromString(peerAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap peer address %s: %w", peerAddr, err)
		}
		if addrInfo.ID == s.host.ID() {
			slog.Warn("Skipping self as bootstrap peer", "peer", addrInfo.ID)
			continue
		}
		bootPeers = append(bootPeers, *addrInfo)
	}
	return bootPeers, nil
}

func (s *discService) initMDNS() error {
	if !s.cfg.Discovery.EnableMDNS {
		slog.Info("mDNS disabled by configuration")
		return nil
	}
	mdnsService := mdns.NewMdnsService(s.host, s.cfg.Discovery.MDNSServiceName, s)
	s.mdns = mdnsService
	slog.Info("mDNS initialized", "service-name", s.cfg.Discovery.MDNSServiceName)
	return nil
}

func (s *discService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return fmt.Errorf("discovery service already started")
	}
	slog.Info("Starting discovery service...")

	if s.dht != nil {
		if err := s.dht.Bootstrap(s.ctx); err != nil {
			return fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
		slog.Info("DHT bootstrap initiated")
		go s.maintainConnections(s.ctx)
	}
	if s.mdns != nil {
		if err := s.mdns.Start(); err != nil {
			return fmt.Errorf("failed to start mDNS: %w", err)
		}
		slog.Info("mDNS discovery started")
	}
	s.started = true
	slog.Info("Discovery service started successfully")
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
				if err := s.connectToPeer(peerInfo); err != nil {
					slog.Debug("Failed to connect to bootstrap peer", "peer", peerInfo.ID, "error", err)
				} else {
					slog.Debug("Connected to bootstrap peer", "peer", peerInfo.ID)
				}
			}
			s.mu.RUnlock()

			if s.dht != nil && s.cfg.Discovery.AdvertiseServiceName != "" {
				if s.lastAdvertisement.IsZero() || time.Since(s.lastAdvertisement) >= s.cfg.Discovery.AdvertiseInterval {
					if err := s.Advertise(ctx, s.cfg.Discovery.AdvertiseServiceName); err != nil {
						slog.Error("Failed to advertise in DHT", "error", err)
					} else {
						s.lastAdvertisement = time.Now()
						slog.Debug("Successfully advertised in DHT", "peer", s.host.ID())
					}
				}

				findCtx, findCancel := context.WithCancel(ctx)
				peers, err := s.rdiscov.FindPeers(findCtx, s.cfg.Discovery.AdvertiseServiceName)
				if err != nil {
					findCancel()
					slog.Error("Failed to find closest peers", "error", err)
					continue
				}
				count := s.cfg.Discovery.MaxDHTReconnectPeerCount
				for peerInfo := range peers {
					if s.host.ID() == peerInfo.ID {
						continue
					}
					if err := s.connectToPeer(peerInfo); err != nil {
						slog.Warn("Failed to connect to DHT peer", "peer", peerInfo.ID, "error", err)
					} else {
						slog.Debug("Connected to DHT peer", "peer", peerInfo.ID)
						count--
					}
					if count <= 0 {
						slog.Debug("Reached connection limit for DHT peers", "limit", count)
						break
					}
				}
				findCancel()
			}
		}
	}
}

func (s *discService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return nil
	}
	slog.Info("Stopping discovery service...")
	s.cancel()
	if s.dht != nil {
		if err := s.dht.Close(); err != nil {
			slog.Error("Error closing DHT", "error", err)
		}
	}
	if s.mdns != nil {
		if err := s.mdns.Close(); err != nil {
			slog.Error("Error closing mDNS", "error", err)
		}
	}
	s.started = false
	slog.Info("Discovery service stopped successfully")
	return nil
}

func (s *discService) GetDiscoveryClient() *drouting.RoutingDiscovery {
	if s.rdiscov == nil {
		slog.Warn("Routing discovery client is not initialized")
		return nil
	}
	slog.Debug("Returning routing discovery client")
	return s.rdiscov
}

func (s *discService) Advertise(ctx context.Context, topic string) error {
	if s.rdiscov == nil {
		return fmt.Errorf("routing discovery not available")
	}
	slog.Debug("Advertising for topic", "topic", topic)
	ttl := s.cfg.Discovery.AdvertiseTTL
	if ttl == 0 {
		ttl = 12 * time.Hour
	}
	_, err := s.rdiscov.Advertise(ctx, topic, discovery.TTL(ttl))
	if err != nil {
		return fmt.Errorf("failed to advertise for topic %s: %w", topic, err)
	}
	slog.Debug("Successfully advertised for topic", "topic", topic, "ttl", ttl)
	return nil
}

// HandlePeerFound processes a newly discovered mDNS peer and attempts to connect.
func (s *discService) HandlePeerFound(peerInfo peer.AddrInfo) {
	slog.Debug("Discovered new mDNS peer", "peer", peerInfo.ID, "addresses", peerInfo.Addrs)

	// Attempt to connect to the discovered peer
	if err := s.connectToPeer(peerInfo); err != nil {
		slog.Warn("Failed to connect to mDNS peer", "peer", peerInfo.ID, "error", err)
	}
}

func (s *discService) connectToPeer(peerInfo peer.AddrInfo) error {
	if s.host.Network().Connectedness(peerInfo.ID) == network.Connected {
		slog.Debug("Already connected to peer", "peer", peerInfo.ID)
		return nil
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.cfg.Discovery.ConnectionTimeout)
	defer cancel()
	return s.host.Connect(ctx, peerInfo)
}
