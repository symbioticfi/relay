package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"

	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

// TestService_IntegrationSuccessful tests full P2P communication between two services
func TestService_IntegrationSuccessful(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	// Create two libp2p hosts
	service1 := createTestService(t, false, nil)
	service2 := createTestService(t, false, nil)

	// Connect the peers
	host1Addr := host.InfoFromHost(service1.host)
	err := service2.addPeer(*host1Addr)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Small delay to ensure the gossip protocol is set up

	// Wait for connection to establish
	require.Eventually(t, func() bool {
		return len(service1.host.Peerstore().Peers()) > 0 && len(service2.host.Peerstore().Peers()) > 0
	}, time.Second, time.Millisecond*100)

	// Set up message listener on service2
	var receivedMsg p2pEntity.P2PMessage[symbiotic.Signature]

	done := make(chan struct{})
	require.NoError(t, service2.StartSignatureMessageListener(func(ctx context.Context, msg p2pEntity.P2PMessage[symbiotic.Signature]) error {
		receivedMsg = msg
		close(done)
		return nil
	}))

	priv, err := symbioticCrypto.GeneratePrivateKey(symbiotic.KeyTag(1).Type())
	require.NoError(t, err)

	// Prepare test message
	testSignatureMsg := symbiotic.Signature{
		KeyTag:      symbiotic.KeyTag(1),
		Epoch:       symbiotic.Epoch(123),
		MessageHash: symbiotic.RawMessageHash("test message hash"),
		Signature:   symbiotic.RawSignature("test signature"),
		PublicKey:   priv.PublicKey(),
	}

	// Send the message from service1
	err = service1.BroadcastSignatureGeneratedMessage(ctx, testSignatureMsg)
	require.NoError(t, err)

	select {
	case <-done:
		// Verify the received message
		assert.Equal(t, service1.host.ID().String(), receivedMsg.SenderInfo.Sender)
		assert.NotNil(t, receivedMsg.SenderInfo.PublicKey)
		assert.Equal(t, testSignatureMsg.KeyTag, receivedMsg.Message.KeyTag)
		assert.Equal(t, testSignatureMsg.Epoch, receivedMsg.Message.Epoch)
		assert.Equal(t, testSignatureMsg.MessageHash, receivedMsg.Message.MessageHash)
		assert.Equal(t, testSignatureMsg.Signature, receivedMsg.Message.Signature)
		assert.Equal(t, testSignatureMsg.PublicKey, receivedMsg.Message.PublicKey)
	case <-ctx.Done():
		require.Fail(t, "Test timed out waiting for message")
	}
}

// TestService_IntegrationFailedSignature tests P2P communication with a message that fails signature verification
func TestService_IntegrationFailedSignature(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	tr := &rejectTracer{
		rejectCh: make(chan *pubsub_pb.TraceEvent, 1),
	}
	defer close(tr.rejectCh)
	// Create two libp2p hosts
	service1 := createTestService(t, false, nil)
	service2 := createTestService(t, true, tr)

	// Connect the peers
	host1Addr := host.InfoFromHost(service1.host)
	err := service2.addPeer(*host1Addr)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Small delay to ensure the gossip protocol is set up

	// Wait for connection to establish
	require.Eventually(t, func() bool {
		return len(service1.host.Peerstore().Peers()) > 0 && len(service2.host.Peerstore().Peers()) > 0
	}, time.Second, time.Millisecond*100)

	priv, err := symbioticCrypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	// Send the message from service1
	err = service1.BroadcastSignatureGeneratedMessage(ctx, symbiotic.Signature{
		PublicKey: priv.PublicKey(),
	})
	require.NoError(t, err)

	select {
	case evt := <-tr.rejectCh:
		require.Equal(t, topicSignatureReady, lo.FromPtr(evt.RejectMessage.Topic))
		require.Equal(t, "unexpected signature", lo.FromPtr(evt.RejectMessage.Reason))
	case <-ctx.Done():
		require.Fail(t, "Test timed out waiting for message")
	}
}

func createTestService(t *testing.T, skipMessageSigning bool, tracer pubsub.EventTracer) *Service {
	t.Helper()

	p2pIdentityPKRaw, err := symbioticCrypto.GeneratePrivateKey(symbiotic.KeyTypeEcdsaSecp256k1)
	require.NoError(t, err)

	p2pIdentityPK, err := crypto.UnmarshalSecp256k1PrivateKey(p2pIdentityPKRaw.Bytes())
	require.NoError(t, err)

	h, err := libp2p.New(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Identity(p2pIdentityPK),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultMuxers,
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, h.Close())
	})

	service, err := NewService(t.Context(), Config{
		Host:            h,
		SkipMessageSign: skipMessageSigning,
		Metrics:         &mockMetrics{},
		Discovery:       DefaultDiscoveryConfig(),
		EventTracer:     tracer,
		Handler:         myHandler{},
	}, signals.Config{
		BufferSize:  5,
		WorkerCount: 1,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, service.Close())
	})

	return service
}

type rejectTracer struct {
	rejectCh chan *pubsub_pb.TraceEvent
}

func (rt *rejectTracer) Trace(evt *pubsub_pb.TraceEvent) {
	if *evt.Type != pubsub_pb.TraceEvent_REJECT_MESSAGE {
		return
	}
	rt.rejectCh <- evt
}

// mockMetrics implements the metrics interface for testing
type mockMetrics struct{}

func (m *mockMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

func (m *mockMetrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, stream)
	}
}

func (m *mockMetrics) ObserveP2PPeerMessageSent(messageType, status string) {}
