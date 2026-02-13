package votingpower

import (
	"context"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	votingpowerv1 "github.com/symbioticfi/relay/internal/gen/api/votingpower/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type testServer struct {
	votingpowerv1.UnimplementedVotingPowerProviderServiceServer

	mu     sync.Mutex
	calls  int
	lastMD metadata.MD

	fn func(ctx context.Context, req *votingpowerv1.GetVotingPowersAtRequest, call int) (*votingpowerv1.GetVotingPowersAtResponse, error)
}

func (s *testServer) GetVotingPowersAt(ctx context.Context, req *votingpowerv1.GetVotingPowersAtRequest) (*votingpowerv1.GetVotingPowersAtResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	call := s.calls
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		s.lastMD = md.Copy()
	}
	if s.fn == nil {
		return &votingpowerv1.GetVotingPowersAtResponse{}, nil
	}
	return s.fn(ctx, req, call)
}

func (s *testServer) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *testServer) metadata() metadata.MD {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastMD
}

func startTestServer(t *testing.T, srv *testServer) (string, *health.Server) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	votingpowerv1.RegisterVotingPowerProviderServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	t.Cleanup(func() {
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		grpcServer.Stop()
		_ = listener.Close()
	})

	return listener.Addr().String(), healthServer
}

func testProviderID() ProviderID {
	return ProviderID{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}
}

func providerAddress(id ProviderID) symbiotic.CrossChainAddress {
	var addr common.Address
	copy(addr[:10], id[:])
	return symbiotic.CrossChainAddress{ChainId: 0, Address: addr}
}

func TestNewClient_DuplicateProviderID(t *testing.T) {
	id := testProviderID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewClient(ctx, []ProviderConfig{
		{ID: providerIDString(id), URL: "127.0.0.1:1"},
		{ID: providerIDString(id), URL: "127.0.0.1:2"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate provider id")
}

func TestClient_GetVotingPowers_HappyPathAndSorting(t *testing.T) {
	srv := &testServer{
		fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
			return &votingpowerv1.GetVotingPowersAtResponse{VotingPowers: []*votingpowerv1.OperatorVotingPower{
				{Operator: "0x0000000000000000000000000000000000000002", VotingPower: "0"},
				{Operator: "0x0000000000000000000000000000000000000001", VotingPower: "100"},
			}}, nil
		},
	}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	provider := providerAddress(id)
	result, err := client.GetVotingPowers(ctx, provider, symbiotic.Timestamp(123))
	require.NoError(t, err)
	require.Len(t, result, 2)

	require.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000001"), result[0].Operator)
	require.Equal(t, provider.Address, result[0].Vaults[0].Vault)
	require.Equal(t, 0, result[0].Vaults[0].VotingPower.Cmp(big.NewInt(100)))

	require.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000002"), result[1].Operator)
	require.Equal(t, 0, result[1].Vaults[0].VotingPower.Cmp(big.NewInt(0)))
}

func TestClient_GetVotingPowers_DuplicateOperatorsMerged(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return &votingpowerv1.GetVotingPowersAtResponse{VotingPowers: []*votingpowerv1.OperatorVotingPower{
			{Operator: "0x0000000000000000000000000000000000000001", VotingPower: "7"},
			{Operator: "0x0000000000000000000000000000000000000001", VotingPower: "5"},
		}}, nil
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	ctx := context.Background()
	client, err := NewClient(ctx, []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	result, err := client.GetVotingPowers(ctx, providerAddress(id), 100)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, 0, result[0].Vaults[0].VotingPower.Cmp(big.NewInt(12)))
}

func TestClient_GetVotingPowers_EmptyResponse(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return &votingpowerv1.GetVotingPowersAtResponse{}, nil
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	result, err := client.GetVotingPowers(context.Background(), providerAddress(id), 100)
	require.NoError(t, err)
	require.Len(t, result, 0)
}

func TestClient_GetVotingPowers_InvalidOperator(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return &votingpowerv1.GetVotingPowersAtResponse{VotingPowers: []*votingpowerv1.OperatorVotingPower{{
			Operator:    "invalid",
			VotingPower: "1",
		}}}, nil
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	_, err = client.GetVotingPowers(context.Background(), providerAddress(id), 100)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid operator address")
}

func TestClient_GetVotingPowers_InvalidVotingPower(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return &votingpowerv1.GetVotingPowersAtResponse{VotingPowers: []*votingpowerv1.OperatorVotingPower{{
			Operator:    "0x0000000000000000000000000000000000000001",
			VotingPower: "NaN",
		}}}, nil
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	_, err = client.GetVotingPowers(context.Background(), providerAddress(id), 100)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid voting power")
}

func TestClient_GetVotingPowers_UnknownProvider(t *testing.T) {
	id := testProviderID()
	srv := &testServer{}
	url, _ := startTestServer(t, srv)

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	unknownID := ProviderID{1, 2, 3}
	_, err = client.GetVotingPowers(context.Background(), providerAddress(unknownID), 100)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured")
}

func TestClient_GetVotingPowers_NoRetry(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return nil, context.DeadlineExceeded
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	_, err = client.GetVotingPowers(context.Background(), providerAddress(id), 100)
	require.Error(t, err)
	require.Equal(t, 1, srv.callCount())
}

func TestClient_NewClient_HealthCheckFailure(t *testing.T) {
	srv := &testServer{}
	url, healthServer := startTestServer(t, srv)
	id := testProviderID()

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	_, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not serving")
}

func TestClient_GetVotingPowers_Headers(t *testing.T) {
	srv := &testServer{fn: func(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest, _ int) (*votingpowerv1.GetVotingPowersAtResponse, error) {
		return &votingpowerv1.GetVotingPowersAtResponse{}, nil
	}}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{
		ID:      providerIDString(id),
		URL:     url,
		Headers: map[string]string{"authorization": "Bearer test"},
	}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	_, err = client.GetVotingPowers(context.Background(), providerAddress(id), 100)
	require.NoError(t, err)

	md := srv.metadata()
	require.Equal(t, "Bearer test", md.Get("authorization")[0])
}

func TestClient_Close(t *testing.T) {
	srv := &testServer{}
	url, _ := startTestServer(t, srv)
	id := testProviderID()

	client, err := NewClient(context.Background(), []ProviderConfig{{ID: providerIDString(id), URL: url}})
	require.NoError(t, err)
	require.NoError(t, client.Close())
}
