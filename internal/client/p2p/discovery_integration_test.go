package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryService_Start_WithDHTEnabled_InitializesDHT(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS:                     false,
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test-service",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.Start(ctx)

	require.NoError(t, err)
	assert.True(t, ds.started)
	assert.NotNil(t, ds.dht)
	assert.NotNil(t, ds.rdiscov)
	assert.NotNil(t, ds.ctx)
	assert.NotNil(t, ds.cancel)

	err = ds.Close(context.Background())
	require.NoError(t, err)
	assert.False(t, ds.started)
}

func TestDiscoveryService_Start_WithDHTDisabled_SkipsDHT(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS: false,
			DHTMode:    "disabled",
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.Start(ctx)

	require.NoError(t, err)
	assert.True(t, ds.started)
	assert.Nil(t, ds.dht)
	assert.Nil(t, ds.rdiscov)

	err = ds.Close(context.Background())
	require.NoError(t, err)
}

func TestDiscoveryService_Start_WhenAlreadyStarted_ReturnsError(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS: false,
			DHTMode:    "disabled",
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.Start(ctx)
	require.NoError(t, err)

	err = ds.Start(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	ds.Close(context.Background())
}

func TestDiscoveryService_initDHT_WithClientMode_Success(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.initDHT(ctx)

	require.NoError(t, err)
	assert.NotNil(t, ds.dht)
	assert.NotNil(t, ds.rdiscov)
	assert.Empty(t, ds.bootstrapPeers)

	ds.dht.Close()
}

func TestDiscoveryService_initDHT_WithBootstrapPeers_ParsesPeers(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			DHTMode: "client",
			BootstrapPeers: []string{
				"/ip4/127.0.0.1/tcp/4001/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
			},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.initDHT(ctx)

	require.NoError(t, err)
	assert.Len(t, ds.bootstrapPeers, 1)
	assert.Equal(t, "QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt", ds.bootstrapPeers[0].ID.String())

	ds.dht.Close()
}

func TestDiscoveryService_initMDNS_WhenDisabled_ReturnsNil(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS: false,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	err = ds.initMDNS(context.Background())

	require.NoError(t, err)
	assert.Nil(t, ds.mdns)
}

func TestDiscoveryService_initMDNS_WhenEnabled_InitializesMDNS(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS:      true,
			MDNSServiceName: "test-mdns",
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	err = ds.initMDNS(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, ds.mdns)

	err = ds.mdns.Close()
	require.NoError(t, err)
}

func TestDiscoveryService_Close_WithDHTAndMDNS_ClosesAll(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			EnableMDNS:                     true,
			MDNSServiceName:                "test-mdns",
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.Start(ctx)
	require.NoError(t, err)

	err = ds.Close(context.Background())

	require.NoError(t, err)
	assert.False(t, ds.started)
}

func TestDiscoveryService_HandlePeerFound_ConnectsToPeer(t *testing.T) {
	host1, err := libp2p.New()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := libp2p.New()
	require.NoError(t, err)
	defer host2.Close()

	cfg := Config{
		Host: host1,
		Discovery: DiscoveryConfig{
			ConnectionTimeout: 5 * time.Second,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ds.ctx = ctx

	peerInfo := *host.InfoFromHost(host2)

	ds.HandlePeerFound(peerInfo)

	time.Sleep(100 * time.Millisecond)

	connectedness := host1.Network().Connectedness(host2.ID())
	assert.Equal(t, network.Connected, connectedness)
}

func TestDiscoveryService_Advertise_WithValidRdiscov_SuccessOrExpectedError(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ds.initDHT(ctx)
	require.NoError(t, err)
	defer ds.dht.Close()

	err = ds.Advertise(ctx, "test-topic")

	if err != nil {
		assert.Contains(t, err.Error(), "failed to advertise")
	}
}

func TestDiscoveryService_Advertise_WithZeroTTL_CallsAdvertise(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   0,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ds.initDHT(ctx)
	require.NoError(t, err)
	defer ds.dht.Close()

	err = ds.Advertise(ctx, "test-topic")

	if err != nil {
		assert.Contains(t, err.Error(), "failed to advertise")
	}
}

func TestDiscoveryService_GetDiscoveryClient_WhenInitialized_ReturnsClient(t *testing.T) {
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	cfg := Config{
		Host: h,
		Discovery: DiscoveryConfig{
			DHTMode:                        "client",
			BootstrapPeers:                 []string{},
			AdvertiseTTL:                   time.Minute,
			AdvertiseServiceName:           "test",
			AdvertiseInterval:              time.Minute,
			ConnectionTimeout:              5 * time.Second,
			MaxDHTReconnectPeerCount:       10,
			DHTPeerDiscoveryInterval:       time.Minute,
			DHTRoutingTableRefreshInterval: time.Minute,
		},
	}

	ds, err := NewDiscoveryService(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ds.Start(ctx)
	require.NoError(t, err)
	defer ds.Close(context.Background())

	client := ds.GetDiscoveryClient(context.Background())

	assert.NotNil(t, client)
	assert.Equal(t, ds.rdiscov, client)
}
