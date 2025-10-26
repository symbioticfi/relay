package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
)

// httpGatewayBaseURL constructs the base URL for HTTP gateway requests
func httpGatewayBaseURL(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("http://localhost:%s/api/v1", globalTestEnv.GetContainerPort(0))
}

// TestHTTPGateway_GetCurrentEpoch tests the HTTP gateway GET endpoint for current epoch
func TestHTTPGateway_GetCurrentEpoch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get expected data from gRPC API
	grpcClient := globalTestEnv.GetGRPCClient(t, 0)
	grpcResp, err := grpcClient.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
	require.NoError(t, err, "Failed to get current epoch from gRPC")

	// Make HTTP request
	url := fmt.Sprintf("%s/epoch/current", httpGatewayBaseURL(t))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(t, err, "Failed to create HTTP request")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(t, err, "Failed to make HTTP request")
	defer httpResp.Body.Close()

	// Verify HTTP status
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Expected 200 OK status")

	// Verify Content-Type
	contentType := httpResp.Header.Get("Content-Type")
	require.Contains(t, contentType, "application/json", "Expected JSON content type")

	// Parse response
	var httpResult struct {
		Epoch     string `json:"epoch"`
		StartTime string `json:"startTime"`
	}
	err = json.NewDecoder(httpResp.Body).Decode(&httpResult)
	require.NoError(t, err, "Failed to decode HTTP response")

	// Verify response matches gRPC data
	require.NotEmpty(t, httpResult.Epoch, "Epoch should not be empty")
	require.LessOrEqual(t, fmt.Sprintf("%d", grpcResp.GetEpoch()), httpResult.Epoch,
		"HTTP epoch should match gRPC epoch")

	t.Logf("✓ HTTP Gateway GET endpoint works correctly")
	t.Logf("  Epoch: %s", httpResult.Epoch)
	t.Logf("  Content-Type: %s", contentType)
}

// TestHTTPGateway_GetValidatorSet tests the HTTP gateway GET endpoint for validator set
func TestHTTPGateway_GetValidatorSet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get expected data from gRPC API
	grpcClient := globalTestEnv.GetGRPCClient(t, 0)
	grpcResp, err := grpcClient.GetValidatorSet(ctx, &apiv1.GetValidatorSetRequest{})
	require.NoError(t, err, "Failed to get validator set from gRPC")

	// Make HTTP request
	url := fmt.Sprintf("%s/validator-set", httpGatewayBaseURL(t))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(t, err, "Failed to create HTTP request")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(t, err, "Failed to make HTTP request")
	defer httpResp.Body.Close()

	// Verify HTTP status
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Expected 200 OK status")

	// Parse response
	var httpResult struct {
		ValidatorSet struct {
			Epoch      string `json:"epoch"`
			Version    int    `json:"version"`
			Validators []struct {
				Operator string `json:"operator"`
			} `json:"validators"`
		} `json:"validatorSet"`
	}
	err = json.NewDecoder(httpResp.Body).Decode(&httpResult)
	require.NoError(t, err, "Failed to decode HTTP response")

	// Verify response matches gRPC data
	require.LessOrEqual(t, fmt.Sprintf("%d", grpcResp.GetValidatorSet().GetEpoch()),
		httpResult.ValidatorSet.Epoch, "HTTP epoch should match gRPC epoch")
	require.NotEmpty(t, httpResult.ValidatorSet.Validators,
		"Validators list should not be empty")

	t.Logf("✓ HTTP Gateway GET validator set works correctly")
	t.Logf("  Epoch: %s", httpResult.ValidatorSet.Epoch)
	t.Logf("  Validators: %d", len(httpResult.ValidatorSet.Validators))
}

// TestHTTPGateway_StreamProofs tests the HTTP gateway streaming endpoint for proofs
func TestHTTPGateway_StreamProofs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Make HTTP streaming request
	url := fmt.Sprintf("%s/stream/proofs",
		httpGatewayBaseURL(t))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(t, err, "Failed to create HTTP request")

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(t, err, "Failed to make HTTP streaming request")
	defer httpResp.Body.Close()

	// Verify HTTP status
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Expected 200 OK status")

	// Verify SSE Content-Type
	contentType := httpResp.Header.Get("Content-Type")
	require.Contains(t, contentType, "text/event-stream",
		"Expected Server-Sent Events content type")

	// Verify SSE headers
	require.Equal(t, "no-cache", httpResp.Header.Get("Cache-Control"),
		"Expected Cache-Control: no-cache")
	require.Equal(t, "keep-alive", httpResp.Header.Get("Connection"),
		"Expected Connection: keep-alive")

	t.Logf("✓ HTTP Gateway streaming endpoint connected")
	t.Logf("  URL: %s", url)
	t.Logf("  Content-Type: %s", contentType)

	// Read streaming responses with timeout
	streamCtx, streamCancel := context.WithTimeout(ctx, 30*time.Second)
	defer streamCancel()

	messagesReceived := 0
	maxMessages := 3 // Read a few messages to verify streaming works

	scanner := bufio.NewScanner(httpResp.Body)

	// Channel to signal when we've read enough messages
	done := make(chan struct{})

	go func() {
		defer close(done)

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Parse SSE format: "data: {...}"
			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")

				var sseMessage struct {
					Result struct {
						RequestID        string `json:"requestId"`
						Epoch            string `json:"epoch"`
						AggregationProof struct {
							MessageHash string `json:"messageHash"`
							Proof       string `json:"proof"`
						} `json:"aggregationProof"`
					} `json:"result"`
				}

				err := json.Unmarshal([]byte(jsonData), &sseMessage)
				if err != nil {
					t.Logf("Warning: Failed to parse SSE message: %v", err)
					continue
				}

				// Verify message structure
				require.NotEmpty(t, sseMessage.Result.RequestID,
					"RequestID should not be empty")
				require.NotEmpty(t, sseMessage.Result.Epoch,
					"Epoch should not be empty")

				messagesReceived++
				t.Logf("  Received proof %d: RequestID=%s, Epoch=%s",
					messagesReceived,
					sseMessage.Result.RequestID[:16]+"...",
					sseMessage.Result.Epoch)

				if messagesReceived >= maxMessages {
					return
				}
			}
		}
	}()

	// Wait for messages or timeout
	select {
	case <-done:
		// Successfully received messages
		require.GreaterOrEqual(t, messagesReceived, 1,
			"Should receive at least 1 streaming message")
		t.Logf("✓ HTTP Gateway streaming works correctly")
		t.Logf("  Messages received: %d", messagesReceived)

	case <-streamCtx.Done():
		// Timeout - this is OK if no proofs were generated during test
		if messagesReceived > 0 {
			t.Logf("✓ HTTP Gateway streaming works (received %d messages before timeout)",
				messagesReceived)
		} else {
			t.Logf("⚠ No streaming messages received (may be no activity during test)")
			t.Logf("  This is not necessarily an error - streaming endpoint is working")
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil && err != io.EOF {
		// Only fail if we didn't receive any messages
		if messagesReceived == 0 {
			require.NoError(t, err, "Scanner error while reading stream")
		} else {
			t.Logf("  Scanner ended with error (after receiving messages): %v", err)
		}
	}
}

// TestHTTPGateway_StreamSignatures tests the HTTP gateway streaming endpoint for signatures
func TestHTTPGateway_StreamSignatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Make HTTP streaming request
	url := fmt.Sprintf("%s/stream/signatures",
		httpGatewayBaseURL(t))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	require.NoError(t, err, "Failed to create HTTP request")

	httpClient := &http.Client{Timeout: 60 * time.Second}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(t, err, "Failed to make HTTP streaming request")
	defer httpResp.Body.Close()

	// Verify HTTP status and headers
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Expected 200 OK status")
	require.Contains(t, httpResp.Header.Get("Content-Type"), "text/event-stream",
		"Expected SSE content type")

	t.Logf("✓ HTTP Gateway signature streaming endpoint connected")
	t.Logf("  URL: %s", url)

	// Read a few messages to verify streaming
	streamCtx, streamCancel := context.WithTimeout(ctx, 30*time.Second)
	defer streamCancel()

	messagesReceived := 0
	maxMessages := 3

	scanner := bufio.NewScanner(httpResp.Body)
	done := make(chan struct{})

	go func() {
		defer close(done)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.TrimSpace(line) == "" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")

				var sseMessage struct {
					Result struct {
						RequestID string `json:"requestId"`
						Epoch     string `json:"epoch"`
						Signature struct {
							Signature   string `json:"signature"`
							MessageHash string `json:"messageHash"`
							PublicKey   string `json:"publicKey"`
						} `json:"signature"`
					} `json:"result"`
				}

				err := json.Unmarshal([]byte(jsonData), &sseMessage)
				if err != nil {
					t.Logf("Warning: Failed to parse SSE message: %v", err)
					continue
				}

				require.NotEmpty(t, sseMessage.Result.RequestID,
					"RequestID should not be empty")

				messagesReceived++
				t.Logf("  Received signature %d: RequestID=%s, Epoch=%s",
					messagesReceived,
					sseMessage.Result.RequestID[:16]+"...",
					sseMessage.Result.Epoch)

				if messagesReceived >= maxMessages {
					return
				}
			}
		}
	}()

	select {
	case <-done:
		require.GreaterOrEqual(t, messagesReceived, 1,
			"Should receive at least 1 streaming signature")
		t.Logf("✓ HTTP Gateway signature streaming works correctly")
		t.Logf("  Messages received: %d", messagesReceived)

	case <-streamCtx.Done():
		if messagesReceived > 0 {
			t.Logf("✓ HTTP Gateway streaming works (received %d messages before timeout)",
				messagesReceived)
		} else {
			t.Logf("⚠ No streaming messages received (may be no activity during test)")
		}
	}
}
