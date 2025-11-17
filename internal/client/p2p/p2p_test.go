package p2p

import (
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate_Success(t *testing.T) {
	host, err := libp2p.New()
	require.NoError(t, err)
	defer host.Close()

	cfg := Config{
		Host:            host,
		SkipMessageSign: false,
		Metrics:         &mockMetrics{},
		Discovery:       DefaultDiscoveryConfig(),
		Handler:         &GRPCHandler{syncHandler: &mockSyncRequestHandler{}},
	}

	err = cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_MissingHost(t *testing.T) {
	cfg := Config{
		Host:            nil,
		SkipMessageSign: false,
		Metrics:         &mockMetrics{},
		Discovery:       DefaultDiscoveryConfig(),
		Handler:         &GRPCHandler{syncHandler: &mockSyncRequestHandler{}},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Host")
}

func TestConfig_Validate_MissingMetrics(t *testing.T) {
	host, err := libp2p.New()
	require.NoError(t, err)
	defer host.Close()

	cfg := Config{
		Host:            host,
		SkipMessageSign: false,
		Metrics:         nil,
		Discovery:       DefaultDiscoveryConfig(),
		Handler:         &GRPCHandler{syncHandler: &mockSyncRequestHandler{}},
	}

	err = cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Metrics")
}

func TestConfig_Validate_MissingHandler(t *testing.T) {
	host, err := libp2p.New()
	require.NoError(t, err)
	defer host.Close()

	cfg := Config{
		Host:            host,
		SkipMessageSign: false,
		Metrics:         &mockMetrics{},
		Discovery:       DefaultDiscoveryConfig(),
		Handler:         nil,
	}

	err = cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Handler")
}

func TestDefaultDiscoveryConfig_ReturnsValidConfig(t *testing.T) {
	cfg := DefaultDiscoveryConfig()

	assert.False(t, cfg.EnableMDNS)
	assert.Equal(t, "symbiotic-mdns", cfg.MDNSServiceName)
	assert.Equal(t, "server", cfg.DHTMode)
	assert.NotZero(t, cfg.AdvertiseTTL)
	assert.Equal(t, "symbiotic-advertise", cfg.AdvertiseServiceName)
	assert.NotZero(t, cfg.AdvertiseInterval)
	assert.NotZero(t, cfg.ConnectionTimeout)
	assert.Equal(t, 20, cfg.MaxDHTReconnectPeerCount)
	assert.NotZero(t, cfg.DHTPeerDiscoveryInterval)
	assert.NotZero(t, cfg.DHTRoutingTableRefreshInterval)
}
