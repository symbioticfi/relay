package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ID_ReturnsHostID(t *testing.T) {
	service := createTestService(t, false, nil)

	serviceID := service.ID()

	require.NotEmpty(t, serviceID)
	assert.Equal(t, service.host.ID().String(), serviceID)
}

func TestService_Broadcast_TopicNotFound_ReturnsError(t *testing.T) {
	service := createTestService(t, false, nil)

	err := service.broadcast(context.Background(), "/nonexistent/topic", []byte("test data"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "topic")
	assert.Contains(t, err.Error(), "not found")
}

func TestService_AddPeer_SkipsSelfConnection(t *testing.T) {
	service := createTestService(t, false, nil)

	selfInfo := service.host.Peerstore().PeerInfo(service.host.ID())

	err := service.addPeer(selfInfo)

	require.NoError(t, err)
}

func TestService_AddPeer_ConnectsToValidPeer(t *testing.T) {
	service1 := createTestService(t, false, nil)
	service2 := createTestService(t, false, nil)

	peer2Info := service2.host.Peerstore().PeerInfo(service2.host.ID())

	err := service1.addPeer(peer2Info)

	require.NoError(t, err)
	assert.Contains(t, service1.host.Network().Peers(), service2.host.ID())
}
