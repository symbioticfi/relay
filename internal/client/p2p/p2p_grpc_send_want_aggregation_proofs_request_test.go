package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestSendWantAggregationProofsRequest_HappyPath(t *testing.T) {
	serverHost, err := libp2p.New()
	require.NoError(t, err)
	defer serverHost.Close()

	clientHost, err := libp2p.New()
	require.NoError(t, err)
	defer clientHost.Close()

	serverHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(serverHost.ID(), serverHost.Addrs(), peerstore.PermanentAddrTTL)

	err = clientHost.Connect(context.Background(), peer.AddrInfo{
		ID:    serverHost.ID(),
		Addrs: serverHost.Addrs(),
	})
	require.NoError(t, err)

	testHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	testHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	request := entity.WantAggregationProofsRequest{
		RequestIDs: []common.Hash{testHash1, testHash2},
	}

	keyTag := symbiotic.KeyTag(15)
	epoch := symbiotic.Epoch(777)

	expectedProof1 := symbiotic.AggregationProof{
		MessageHash: testHash1.Bytes(),
		KeyTag:      keyTag,
		Epoch:       epoch,
		Proof:       []byte("aggregation_proof_1"),
	}
	expectedProof2 := symbiotic.AggregationProof{
		MessageHash: testHash2.Bytes(),
		KeyTag:      keyTag,
		Epoch:       epoch,
		Proof:       []byte("aggregation_proof_2"),
	}

	expectedResponse := entity.WantAggregationProofsResponse{
		Proofs: map[common.Hash]symbiotic.AggregationProof{
			testHash1: expectedProof1,
			testHash2: expectedProof2,
		},
	}

	mockHandler := &mockAggregationProofHandler{
		expectedRequest:  request,
		responseToReturn: expectedResponse,
	}

	serverService, err := NewService(context.Background(), Config{
		Host:            serverHost,
		SkipMessageSign: true,
		Metrics:         &mockMetrics{},
		Discovery: DiscoveryConfig{
			DHTMode:              "client",
			BootstrapPeers:       []string{},
			AdvertiseTTL:         time.Minute,
			AdvertiseServiceName: "test",
			AdvertiseInterval:    time.Second,
		},
		Handler: &GRPCHandler{
			syncHandler: mockHandler,
		},
	}, signals.Config{
		BufferSize:  5,
		WorkerCount: 5,
	})
	require.NoError(t, err)

	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		err := serverService.StartGRPCServer(serverCtx)
		if err != nil && serverCtx.Err() == nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	defer func() {
		serverCancel()
		select {
		case <-serverDone:
		case <-time.After(2 * time.Second):
			require.Fail(t, "Server did not shut down within timeout")
		}
	}()

	clientService, err := NewService(context.Background(), Config{
		Host:            clientHost,
		SkipMessageSign: true,
		Metrics:         &mockMetrics{},
		Discovery: DiscoveryConfig{
			EnableMDNS:           false,
			MDNSServiceName:      "",
			DHTMode:              "client",
			BootstrapPeers:       []string{},
			AdvertiseTTL:         time.Minute,
			AdvertiseServiceName: "test",
			AdvertiseInterval:    time.Second,
		},
		EventTracer: nil,
		Handler:     NewP2PHandler(mockHandler),
	}, signals.Config{
		BufferSize:  5,
		WorkerCount: 5,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := clientService.SendWantAggregationProofsRequest(ctx, request)
	require.NoError(t, err)

	require.NotNil(t, response)
	require.Len(t, response.Proofs, len(expectedResponse.Proofs))

	proof1, exists := response.Proofs[testHash1]
	require.True(t, exists)
	require.Equal(t, expectedProof1, proof1)

	proof2, exists := response.Proofs[testHash2]
	require.True(t, exists)
	require.Equal(t, expectedProof2, proof2)

	require.True(t, mockHandler.wasCalled)
	require.Len(t, mockHandler.receivedRequest.RequestIDs, 2)
	require.Contains(t, mockHandler.receivedRequest.RequestIDs, testHash1)
	require.Contains(t, mockHandler.receivedRequest.RequestIDs, testHash2)
}

type mockAggregationProofHandler struct {
	expectedRequest  entity.WantAggregationProofsRequest
	responseToReturn entity.WantAggregationProofsResponse
	wasCalled        bool
	receivedRequest  entity.WantAggregationProofsRequest
}

func (m *mockAggregationProofHandler) HandleWantSignaturesRequest(_ context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	return entity.WantSignaturesResponse{
		Signatures: make(map[common.Hash][]entity.ValidatorSignature),
	}, nil
}

func (m *mockAggregationProofHandler) HandleWantAggregationProofsRequest(_ context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	m.wasCalled = true
	m.receivedRequest = request
	return m.responseToReturn, nil
}
