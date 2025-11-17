package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockHost(t *testing.T) host.Host {
	t.Helper()

	h, err := libp2p.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, h.Close())
	})

	return h
}

func TestDiscoveryService_determineDHTMode_Client(t *testing.T) {
	ds := &DiscoveryService{
		cfg: Config{
			Discovery: DiscoveryConfig{
				DHTMode: "client",
			},
		},
	}

	mode := ds.determineDHTMode(context.Background())
	assert.Equal(t, dht.ModeClient, mode)
}

func TestDiscoveryService_determineDHTMode_Server(t *testing.T) {
	ds := &DiscoveryService{
		cfg: Config{
			Discovery: DiscoveryConfig{
				DHTMode: "server",
			},
		},
	}

	mode := ds.determineDHTMode(context.Background())
	assert.Equal(t, dht.ModeServer, mode)
}

func TestDiscoveryService_determineDHTMode_Auto(t *testing.T) {
	ds := &DiscoveryService{
		cfg: Config{
			Discovery: DiscoveryConfig{
				DHTMode: "auto",
			},
		},
	}

	mode := ds.determineDHTMode(context.Background())
	assert.Equal(t, dht.ModeAuto, mode)
}

func TestDiscoveryService_determineDHTMode_Empty(t *testing.T) {
	ds := &DiscoveryService{
		cfg: Config{
			Discovery: DiscoveryConfig{
				DHTMode: "",
			},
		},
	}

	mode := ds.determineDHTMode(context.Background())
	assert.Equal(t, dht.ModeAuto, mode)
}

func TestDiscoveryService_determineDHTMode_Invalid(t *testing.T) {
	ds := &DiscoveryService{
		cfg: Config{
			Discovery: DiscoveryConfig{
				DHTMode: "invalid-mode",
			},
		},
	}

	mode := ds.determineDHTMode(context.Background())
	assert.Equal(t, dht.ModeAuto, mode)
}

func TestDiscoveryService_parseBootstrapPeers_EmptyList(t *testing.T) {
	host := createMockHost(t)
	ds := &DiscoveryService{
		host: host,
		cfg: Config{
			Discovery: DiscoveryConfig{
				BootstrapPeers: []string{},
			},
		},
	}

	peers, err := ds.parseBootstrapPeers(context.Background())
	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestDiscoveryService_parseBootstrapPeers_ValidPeer(t *testing.T) {
	host := createMockHost(t)
	ds := &DiscoveryService{
		host: host,
		cfg: Config{
			Discovery: DiscoveryConfig{
				BootstrapPeers: []string{
					"/ip4/127.0.0.1/tcp/4001/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
				},
			},
		},
	}

	peers, err := ds.parseBootstrapPeers(context.Background())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	assert.Equal(t, "QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt", peers[0].ID.String())
}

func TestDiscoveryService_parseBootstrapPeers_InvalidPeer(t *testing.T) {
	host := createMockHost(t)
	ds := &DiscoveryService{
		host: host,
		cfg: Config{
			Discovery: DiscoveryConfig{
				BootstrapPeers: []string{
					"invalid-peer-address",
				},
			},
		},
	}

	_, err := ds.parseBootstrapPeers(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid bootstrap peer address")
}

func TestDiscoveryService_parseBootstrapPeers_SkipsSelfPeer(t *testing.T) {
	host := createMockHost(t)
	selfPeerAddr := "/ip4/127.0.0.1/tcp/4001/p2p/" + host.ID().String()

	ds := &DiscoveryService{
		host: host,
		cfg: Config{
			Discovery: DiscoveryConfig{
				BootstrapPeers: []string{
					selfPeerAddr,
					"/ip4/127.0.0.1/tcp/4002/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
				},
			},
		},
	}

	peers, err := ds.parseBootstrapPeers(context.Background())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	assert.Equal(t, "QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt", peers[0].ID.String())
}

func TestDiscoveryService_parseBootstrapPeers_MultiplePeers(t *testing.T) {
	host := createMockHost(t)
	ds := &DiscoveryService{
		host: host,
		cfg: Config{
			Discovery: DiscoveryConfig{
				BootstrapPeers: []string{
					"/ip4/127.0.0.1/tcp/4001/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
					"/ip4/127.0.0.1/tcp/4002/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
				},
			},
		},
	}

	peers, err := ds.parseBootstrapPeers(context.Background())
	require.NoError(t, err)
	require.Len(t, peers, 2)

	peerIDs := make([]string, len(peers))
	for i, p := range peers {
		peerIDs[i] = p.ID.String()
	}

	assert.Contains(t, peerIDs, "QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt")
	assert.Contains(t, peerIDs, "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
}

func TestDiscoveryService_GetDiscoveryClient_WhenNotInitialized_ReturnsNil(t *testing.T) {
	host := createMockHost(t)

	ds := &DiscoveryService{
		host:    host,
		rdiscov: nil,
	}

	client := ds.GetDiscoveryClient(context.Background())
	assert.Nil(t, client)
}

func TestNewDiscoveryService_WithValidHost_Success(t *testing.T) {
	host := createMockHost(t)

	cfg := Config{
		Host:      host,
		Discovery: DefaultDiscoveryConfig(),
	}

	ds, err := NewDiscoveryService(cfg)

	require.NoError(t, err)
	require.NotNil(t, ds)
	assert.Equal(t, host, ds.host)
	assert.Equal(t, cfg, ds.cfg)
	assert.NotNil(t, ds.bootstrapPeers)
	assert.Empty(t, ds.bootstrapPeers)
}

func TestNewDiscoveryService_WithNilHost_ReturnsError(t *testing.T) {
	cfg := Config{
		Host:      nil,
		Discovery: DefaultDiscoveryConfig(),
	}

	ds, err := NewDiscoveryService(cfg)

	require.Error(t, err)
	assert.Nil(t, ds)
	assert.Contains(t, err.Error(), "host cannot be nil")
}

func TestDiscoveryService_Advertise_WhenRdiscovIsNil_ReturnsError(t *testing.T) {
	host := createMockHost(t)

	ds := &DiscoveryService{
		host:    host,
		rdiscov: nil,
	}

	err := ds.Advertise(context.Background(), "test-topic")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "routing discovery not available")
}

func TestDiscoveryService_connectToPeer_AlreadyConnected_ReturnsNil(t *testing.T) {
	host1 := createMockHost(t)
	host2 := createMockHost(t)

	err := host1.Connect(context.Background(), *host.InfoFromHost(host2))
	require.NoError(t, err)

	ds := &DiscoveryService{
		host: host1,
		cfg: Config{
			Discovery: DiscoveryConfig{
				ConnectionTimeout: 5 * time.Second,
			},
		},
	}

	peerInfo := *host.InfoFromHost(host2)
	err = ds.connectToPeer(context.Background(), peerInfo)

	require.NoError(t, err)
}

func TestDiscoveryService_connectToPeer_NotConnected_ConnectsToPeer(t *testing.T) {
	host1 := createMockHost(t)
	host2 := createMockHost(t)

	ds := &DiscoveryService{
		host: host1,
		cfg: Config{
			Discovery: DiscoveryConfig{
				ConnectionTimeout: 5 * time.Second,
			},
		},
	}

	peerInfo := *host.InfoFromHost(host2)
	err := ds.connectToPeer(context.Background(), peerInfo)

	require.NoError(t, err)
	assert.Equal(t, network.Connected, host1.Network().Connectedness(host2.ID()))
}

func TestDiscoveryService_connectToPeer_WithTimeout_ReturnsError(t *testing.T) {
	host1 := createMockHost(t)

	ds := &DiscoveryService{
		host: host1,
		cfg: Config{
			Discovery: DiscoveryConfig{
				ConnectionTimeout: 1 * time.Nanosecond,
			},
		},
	}

	peerInfo := peer.AddrInfo{
		ID:    "QmNonExistentPeer",
		Addrs: []multiaddr.Multiaddr{},
	}

	err := ds.connectToPeer(context.Background(), peerInfo)

	require.Error(t, err)
}

func TestDiscoveryService_Close_WhenNotStarted_ReturnsNil(t *testing.T) {
	host := createMockHost(t)

	ds := &DiscoveryService{
		host:    host,
		started: false,
	}

	err := ds.Close(context.Background())

	require.NoError(t, err)
}
