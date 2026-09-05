package api_server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"google.golang.org/grpc"
)

type regressionSigningServer struct {
	apiv1.UnimplementedSymbioticAPIServiceServer

	calls atomic.Int32
}

func (s *regressionSigningServer) SignMessage(context.Context, *apiv1.SignMessageRequest) (*apiv1.SignMessageResponse, error) {
	s.calls.Add(1)
	return &apiv1.SignMessageResponse{RequestId: "accepted"}, nil
}
func TestRegressionCrossOriginPlainTextSigning(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	backend := &regressionSigningServer{}
	apiv1.RegisterSymbioticAPIServiceServer(server, backend)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	mux := http.NewServeMux()
	require.NoError(t, setupHttpProxy(t.Context(), listener.Addr().String(), mux)())
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://relay.example/api/v1/sign", strings.NewReader(`{"keyTag":15,"message":"dGVzdA==","requiredEpoch":"1"}`))
	req.Header.Set("Origin", "http://untrusted.example")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	t.Logf("HTTP status=%d signingCalls=%d CORS=%q body=%s", response.Code, backend.calls.Load(), response.Header().Get("Access-Control-Allow-Origin"), response.Body.String())
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Equal(t, int32(0), backend.calls.Load())
}

func TestAPIRequestOriginAndMediaType(t *testing.T) {
	for _, tc := range []struct {
		name, origin, contentType, fetchSite string
		allowed                              bool
	}{
		{"native", "", "application/json", "", true},
		{"same origin", "http://relay.example", "application/json; charset=utf-8", "same-origin", true},
		{"simple post", "", "text/plain", "", false},
		{"foreign JSON", "http://evil.example", "application/json", "", false},
		{"null origin", "null", "application/json", "", false},
		{"different port", "http://relay.example:8080", "application/json", "", false},
		{"different scheme", "https://relay.example", "application/json", "", false},
		{"cross-site without origin", "", "application/json", "cross-site", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://relay.example/api/v1/sign", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Sec-Fetch-Site", tc.fetchSite)
			require.Equal(t, tc.allowed, allowAPIRequest(httptest.NewRecorder(), req))
		})
	}
}
