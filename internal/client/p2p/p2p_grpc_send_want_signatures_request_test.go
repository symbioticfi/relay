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

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/signals"
)

func TestSendWantSignaturesRequest_HappyPath(t *testing.T) {
	// Setup two libp2p hosts (peer1 as client, peer2 as server)
	serverHost, err := libp2p.New()
	require.NoError(t, err)
	defer serverHost.Close()

	clientHost, err := libp2p.New()
	require.NoError(t, err)
	defer clientHost.Close()

	// Manually connect the peers by adding addresses to peerstore
	serverHost.Peerstore().AddAddrs(clientHost.ID(), clientHost.Addrs(), peerstore.PermanentAddrTTL)
	clientHost.Peerstore().AddAddrs(serverHost.ID(), serverHost.Addrs(), peerstore.PermanentAddrTTL)

	// Actually connect the peers
	err = clientHost.Connect(context.Background(), peer.AddrInfo{
		ID:    serverHost.ID(),
		Addrs: serverHost.Addrs(),
	})
	require.NoError(t, err)

	// Create test data for the request
	testHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	testHash2 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	// Create signature bitmaps (requesting signatures from validators 0, 2, 5)
	wantBitmap1 := entity.NewSignatureBitmapOf(0, 2, 5)
	wantBitmap2 := entity.NewSignatureBitmapOf(1, 3)

	request := entity.WantSignaturesRequest{
		WantSignatures: map[common.Hash]entity.SignatureBitmap{
			testHash1: wantBitmap1,
			testHash2: wantBitmap2,
		},
	}

	// Create expected response data
	expectedSig1 := entity.ValidatorSignature{
		ValidatorIndex: 0,
		Signature: entity.SignatureExtended{
			MessageHash: testHash1.Bytes(),
			Signature:   []byte("signature_for_validator_0"),
			PublicKey:   []byte("public_key_validator_0"),
		},
	}
	expectedSig2 := entity.ValidatorSignature{
		ValidatorIndex: 2,
		Signature: entity.SignatureExtended{
			MessageHash: testHash1.Bytes(),
			Signature:   []byte("signature_for_validator_2"),
			PublicKey:   []byte("public_key_validator_2"),
		},
	}
	expectedSig3 := entity.ValidatorSignature{
		ValidatorIndex: 1,
		Signature: entity.SignatureExtended{
			MessageHash: testHash2.Bytes(),
			Signature:   []byte("signature_for_validator_1"),
			PublicKey:   []byte("public_key_validator_1"),
		},
	}

	expectedResponse := entity.WantSignaturesResponse{
		Signatures: map[common.Hash][]entity.ValidatorSignature{
			testHash1: {expectedSig1, expectedSig2},
			testHash2: {expectedSig3},
		},
	}

	// Setup mock sync request handler for the server
	mockHandler := &mockSyncRequestHandler{
		expectedRequest:  request,
		responseToReturn: expectedResponse,
	}

	// Create server P2P service using NewService constructor
	serverService, err := NewService(context.Background(), Config{
		Host:            serverHost,
		SkipMessageSign: true, // Skip message signing for tests
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

	// Start the gRPC server
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
		// Stop server
		serverCancel()
		select {
		case <-serverDone:
		case <-time.After(2 * time.Second):
			require.Fail(t, "Server did not shut down within timeout")
		}
	}()

	// Create client P2P service using NewService constructor
	clientService, err := NewService(context.Background(), Config{
		Host:            clientHost,
		SkipMessageSign: true, // Skip message signing for tests
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
		Handler:     NewP2PHandler(mockHandler), // Not used for client but required
	}, signals.Config{
		BufferSize:  5,
		WorkerCount: 5,
	})
	require.NoError(t, err)

	// Execute the test: send want signatures request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := clientService.SendWantSignaturesRequest(ctx, request)
	require.NoError(t, err)

	// Verify the response
	require.NotNil(t, response)
	require.Len(t, response.Signatures, len(expectedResponse.Signatures))

	// Verify signatures for testHash1
	signatures1, exists := response.Signatures[testHash1]
	require.True(t, exists)
	require.Len(t, signatures1, 2)

	// Find and verify first signature
	var foundSig1, foundSig2 *entity.ValidatorSignature
	for i := range signatures1 {
		switch signatures1[i].ValidatorIndex {
		case 0:
			foundSig1 = &signatures1[i]
		case 2:
			foundSig2 = &signatures1[i]
		}
	}
	require.NotNil(t, foundSig1)
	require.NotNil(t, foundSig2)
	require.Equal(t, expectedSig1, *foundSig1)
	require.Equal(t, expectedSig2, *foundSig2)

	// Verify signatures for testHash2
	signatures2, exists := response.Signatures[testHash2]
	require.True(t, exists)
	require.Len(t, signatures2, 1)
	require.Equal(t, expectedSig3, signatures2[0])

	// Verify that the mock handler was called with the correct request
	require.True(t, mockHandler.wasCalled)
	require.Equal(t, request.WantSignatures[testHash1].GetCardinality(), mockHandler.receivedRequest.WantSignatures[testHash1].GetCardinality())
	require.Equal(t, request.WantSignatures[testHash2].GetCardinality(), mockHandler.receivedRequest.WantSignatures[testHash2].GetCardinality())
}

// mockSyncRequestHandler implements syncRequestHandler for testing
type mockSyncRequestHandler struct {
	expectedRequest  entity.WantSignaturesRequest
	responseToReturn entity.WantSignaturesResponse
	wasCalled        bool
	receivedRequest  entity.WantSignaturesRequest
}

func (m *mockSyncRequestHandler) HandleWantSignaturesRequest(_ context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	m.wasCalled = true
	m.receivedRequest = request
	return m.responseToReturn, nil
}

func (m *mockSyncRequestHandler) HandleWantAggregationProofsRequest(_ context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	// Return empty response for tests that don't need aggregation proof functionality
	return entity.WantAggregationProofsResponse{
		Proofs: make(map[common.Hash]entity.AggregationProof),
	}, nil
}
