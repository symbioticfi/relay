// Basic usage example for the Symbiotic Relay Go client.
//
// This example demonstrates how to:
// 1. Connect to a Symbiotic Relay server
// 2. Get the current epoch
// 3. Sign a message
// 4. Retrieve aggregation proofs
// 5. Get validator set information
// 6. Get individual signatures
// 7. Stream signatures in real-time
// 8. Stream aggregation proofs in real-time
// 9. Stream validator set changes in real-time

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	client "github.com/symbioticfi/relay/api/client/v1"
)

// RelayClient wraps the Symbiotic client with helpful methods
type RelayClient struct {
	client *client.SymbioticClient
	conn   *grpc.ClientConn
}

// NewRelayClient creates a new client connected to the specified server URL
func NewRelayClient(serverURL string) (*RelayClient, error) {
	// Create gRPC connection
	conn, err := grpc.NewClient(serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Errorf("failed to connect to server: %w", err)
	}

	// Create the symbiotic client
	symbioticClient := client.NewSymbioticClient(conn)

	fmt.Printf("Connected to Symbiotic Relay at %s\n", serverURL)

	return &RelayClient{
		client: symbioticClient,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (rc *RelayClient) Close() error {
	return rc.conn.Close()
}

// GetCurrentEpoch gets the current epoch information
func (rc *RelayClient) GetCurrentEpoch(ctx context.Context) (*client.GetCurrentEpochResponse, error) {
	req := &client.GetCurrentEpochRequest{}
	return rc.client.GetCurrentEpoch(ctx, req)
}

// GetLastAllCommitted gets the last all committed epochs for all chains
func (rc *RelayClient) GetLastAllCommitted(ctx context.Context) (*client.GetLastAllCommittedResponse, error) {
	req := &client.GetLastAllCommittedRequest{}
	return rc.client.GetLastAllCommitted(ctx, req)
}

// SignMessage signs a message using the specified key tag
func (rc *RelayClient) SignMessage(ctx context.Context, keyTag uint32, message []byte, requiredEpoch *uint64) (*client.SignMessageResponse, error) {
	req := &client.SignMessageRequest{
		KeyTag:        keyTag,
		Message:       message,
		RequiredEpoch: requiredEpoch,
	}
	return rc.client.SignMessage(ctx, req)
}

// GetAggregationProof gets aggregation proof for a specific request
func (rc *RelayClient) GetAggregationProof(ctx context.Context, requestID string) (*client.GetAggregationProofResponse, error) {
	req := &client.GetAggregationProofRequest{
		RequestId: requestID,
	}
	return rc.client.GetAggregationProof(ctx, req)
}

// GetSignatures gets individual signatures for a request
func (rc *RelayClient) GetSignatures(ctx context.Context, requestID string) (*client.GetSignaturesResponse, error) {
	req := &client.GetSignaturesRequest{
		RequestId: requestID,
	}
	return rc.client.GetSignatures(ctx, req)
}

// GetValidatorSet gets validator set information
func (rc *RelayClient) GetValidatorSet(ctx context.Context, epoch *uint64) (*client.GetValidatorSetResponse, error) {
	req := &client.GetValidatorSetRequest{
		Epoch: epoch,
	}
	return rc.client.GetValidatorSet(ctx, req)
}

// ListenSignatures streams signatures in real-time
func (rc *RelayClient) ListenSignatures(ctx context.Context, startEpoch *uint64) (grpc.ServerStreamingClient[client.ListenSignaturesResponse], error) {
	req := &client.ListenSignaturesRequest{
		StartEpoch: startEpoch,
	}
	return rc.client.ListenSignatures(ctx, req)
}

// ListenProofs streams aggregation proofs in real-time
func (rc *RelayClient) ListenProofs(ctx context.Context, startEpoch *uint64) (grpc.ServerStreamingClient[client.ListenProofsResponse], error) {
	req := &client.ListenProofsRequest{
		StartEpoch: startEpoch,
	}
	return rc.client.ListenProofs(ctx, req)
}

// ListenValidatorSet streams validator set changes in real-time
func (rc *RelayClient) ListenValidatorSet(ctx context.Context, startEpoch *uint64) (grpc.ServerStreamingClient[client.ListenValidatorSetResponse], error) {
	req := &client.ListenValidatorSetRequest{
		StartEpoch: startEpoch,
	}
	return rc.client.ListenValidatorSet(ctx, req)
}

func main() {
	// Initialize client
	serverURL := os.Getenv("RELAY_SERVER_URL")
	if serverURL == "" {
		serverURL = "localhost:8080"
	}

	relayClient, err := NewRelayClient(serverURL)
	if err != nil {
		log.Fatalf("Failed to create relay client: %v", err)
	}
	defer relayClient.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example 1: Get current epoch
	fmt.Println("=== Getting Current Epoch ===")
	epochResponse, err := relayClient.GetCurrentEpoch(ctx)
	if err != nil {
		log.Printf("Failed to get current epoch: %v", err)
	} else {
		fmt.Printf("Current epoch: %d\n", epochResponse.Epoch)
		if epochResponse.StartTime != nil {
			fmt.Printf("Start time: %v\n", epochResponse.StartTime.AsTime())
		}
	}

	// Example 2: Get suggested epoch
	fmt.Println("\n=== Calculate Last Committed Epoch ===")
	suggestedEpoch := 0
	epochInfos, err := relayClient.GetLastAllCommitted(ctx)
	if err != nil {
		log.Printf("Failed to get last committed epoch: %v", err)
	} else {
		for _, info := range epochInfos.GetEpochInfos() {
			if suggestedEpoch == 0 || int(info.GetLastCommittedEpoch()) < suggestedEpoch {
				suggestedEpoch = int(info.GetLastCommittedEpoch())
			}
		}
	}
	fmt.Printf("Last committed epoch: %d\n", suggestedEpoch)

	// Example 3: Get validator set
	fmt.Println("\n=== Getting Validator Set ===")
	validatorSetResp, err := relayClient.GetValidatorSet(ctx, nil)
	if err != nil {
		log.Printf("Failed to get validator set: %v", err)
	} else {
		validatorSet := validatorSetResp.GetValidatorSet()

		fmt.Printf("Validator set version: %d\n", validatorSet.GetVersion())
		fmt.Printf("Epoch: %d\n", validatorSet.GetEpoch())
		fmt.Printf("Status: %v\n", validatorSet.GetStatus())
		fmt.Printf("Number of validators: %d\n", len(validatorSet.GetValidators()))
		fmt.Printf("Quorum threshold: %s\n", validatorSet.GetQuorumThreshold())

		// Display some validator details
		if len(validatorSet.GetValidators()) > 0 {
			firstValidator := validatorSet.GetValidators()[0]
			fmt.Printf("First validator operator: %s\n", firstValidator.GetOperator())
			fmt.Printf("First validator voting power: %s\n", firstValidator.GetVotingPower())
			fmt.Printf("First validator is active: %t\n", firstValidator.GetIsActive())
			fmt.Printf("First validator keys count: %d\n", len(firstValidator.GetKeys()))
		}
	}

	// Example 4: Sign a message
	fmt.Println("\n=== Signing a Message ===")
	messageToSign := []byte("Hello, Symbiotic!")
	keyTag := uint32(15)

	signResponse, err := relayClient.SignMessage(ctx, keyTag, messageToSign, nil)
	if err != nil {
		log.Printf("Failed to sign message: %v", err)
		return
	}

	fmt.Printf("Request id: %s\n", signResponse.GetRequestId())
	fmt.Printf("Epoch: %d\n", signResponse.Epoch)

	// Example 5: Get aggregation proof (this might fail if signing is not complete)
	fmt.Println("\n=== Getting Aggregation Proof ===")
	proofResponse, err := relayClient.GetAggregationProof(ctx, signResponse.GetRequestId())
	if err != nil {
		fmt.Printf("Could not get aggregation proof yet: %v\n", err)
	} else if proofResponse.GetAggregationProof() != nil {
		proof := proofResponse.GetAggregationProof()
		fmt.Printf("Proof length: %d bytes\n", len(proof.GetProof()))
		fmt.Printf("Message hash length: %d bytes\n", len(proof.GetMessageHash()))
	}

	// Example 6: Get individual signatures
	fmt.Println("\n=== Getting Individual Signatures ===")
	signaturesResponse, err := relayClient.GetSignatures(ctx, signResponse.GetRequestId())
	if err != nil {
		fmt.Printf("Could not get signatures yet: %v\n", err)
	} else {
		fmt.Printf("Number of signatures: %d\n", len(signaturesResponse.GetSignatures()))

		for i, signature := range signaturesResponse.Signatures {
			fmt.Printf("Signature %d:\n", i+1)
			fmt.Printf("  - Signature length: %d bytes\n", len(signature.GetSignature()))
			fmt.Printf("  - Public key length: %d bytes\n", len(signature.GetPublicKey()))
			fmt.Printf("  - Message hash length: %d bytes\n", len(signature.GetMessageHash()))
		}
	}

	// Example 7: Listen to signatures stream
	fmt.Println("\n=== Listening to Signatures Stream ===")
	// Create a new context with a shorter timeout for streaming example
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer streamCancel()

	// Start listening from a specific epoch (optional)
	var startEpoch *uint64
	if epochResponse != nil && epochResponse.GetEpoch() > 0 {
		epoch := epochResponse.GetEpoch() - 1 // Start from previous epoch to get some historical data
		startEpoch = &epoch
		fmt.Printf("Starting stream from epoch: %d\n", *startEpoch)
	}

	sigStream, err := relayClient.ListenSignatures(streamCtx, startEpoch)
	if err != nil {
		log.Printf("Failed to start signatures stream: %v", err)
	} else {
		fmt.Println("Listening for signatures (max 5 events)...")
		count := 0
		maxEvents := 5

		for count < maxEvents {
			resp, err := sigStream.Recv()
			if err != nil {
				fmt.Printf("Stream ended or error: %v\n", err)
				break
			}

			count++
			fmt.Printf("Signature event %d:\n", count)
			fmt.Printf("  - Request ID: %s\n", resp.GetRequestId())
			fmt.Printf("  - Epoch: %d\n", resp.GetEpoch())
			if resp.GetSignature() != nil {
				fmt.Printf("  - Signature length: %d bytes\n", len(resp.GetSignature().GetSignature()))
				fmt.Printf("  - Public key length: %d bytes\n", len(resp.GetSignature().GetPublicKey()))
			}
		}
	}

	// Example 8: Listen to aggregation proofs stream
	fmt.Println("\n=== Listening to Aggregation Proofs Stream ===")
	proofStreamCtx, proofStreamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer proofStreamCancel()

	proofStream, err := relayClient.ListenProofs(proofStreamCtx, startEpoch)
	if err != nil {
		log.Printf("Failed to start proofs stream: %v", err)
	} else {
		fmt.Println("Listening for aggregation proofs (max 5 events)...")
		count := 0
		maxEvents := 5

		for count < maxEvents {
			resp, err := proofStream.Recv()
			if err != nil {
				fmt.Printf("Stream ended or error: %v\n", err)
				break
			}

			count++
			fmt.Printf("Proof event %d:\n", count)
			fmt.Printf("  - Request ID: %s\n", resp.GetRequestId())
			fmt.Printf("  - Epoch: %d\n", resp.GetEpoch())
			if resp.GetAggregationProof() != nil {
				fmt.Printf("  - Proof length: %d bytes\n", len(resp.GetAggregationProof().GetProof()))
				fmt.Printf("  - Message hash length: %d bytes\n", len(resp.GetAggregationProof().GetMessageHash()))
			}
		}
	}

	// Example 9: Listen to validator set changes stream
	fmt.Println("\n=== Listening to Validator Set Changes Stream ===")
	vsStreamCtx, vsStreamCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer vsStreamCancel()

	vsStream, err := relayClient.ListenValidatorSet(vsStreamCtx, startEpoch)
	if err != nil {
		log.Printf("Failed to start validator set stream: %v", err)
	} else {
		fmt.Println("Listening for validator set changes (max 3 events)...")
		count := 0
		maxEvents := 3

		for count < maxEvents {
			resp, err := vsStream.Recv()
			if err != nil {
				fmt.Printf("Stream ended or error: %v\n", err)
				break
			}

			count++
			vs := resp.GetValidatorSet()
			if vs != nil {
				fmt.Printf("Validator Set event %d:\n", count)
				fmt.Printf("  - Epoch: %d\n", vs.GetEpoch())
				fmt.Printf("  - Version: %d\n", vs.GetVersion())
				fmt.Printf("  - Status: %v\n", vs.GetStatus())
				fmt.Printf("  - Validators count: %d\n", len(vs.GetValidators()))
				fmt.Printf("  - Quorum threshold: %s\n", vs.GetQuorumThreshold())
			}
		}
	}

	fmt.Println("\n=== Examples completed ===")
}
