// Basic usage example for the Symbiotic Relay Go client.
//
// This example demonstrates how to:
// 1. Connect to a Symbiotic Relay server
// 2. Get the current epoch
// 3. Sign a message
// 4. Retrieve aggregation proofs
// 5. Get validator set information
// 6. Sign and wait for completion via streaming

package main

import (
	"context"
	"fmt"
	"io"
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
func (rc *RelayClient) GetAggregationProof(ctx context.Context, signatureTargetID string) (*client.GetAggregationProofResponse, error) {
	req := &client.GetAggregationProofRequest{
		SignatureTargetId: signatureTargetID,
	}
	return rc.client.GetAggregationProof(ctx, req)
}

// GetSignatures gets individual signatures for a request
func (rc *RelayClient) GetSignatures(ctx context.Context, signatureTargetID string) (*client.GetSignaturesResponse, error) {
	req := &client.GetSignaturesRequest{
		SignatureTargetId: signatureTargetID,
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

// SignMessageAndWait signs a message and waits for aggregation via streaming response
func (rc *RelayClient) SignMessageAndWait(ctx context.Context, keyTag uint32, message []byte, requiredEpoch *uint64) (grpc.ServerStreamingClient[client.SignMessageWaitResponse], error) {
	req := &client.SignMessageWaitRequest{
		KeyTag:        keyTag,
		Message:       message,
		RequiredEpoch: requiredEpoch,
	}
	return rc.client.SignMessageWait(ctx, req)
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
		for _, info := range epochInfos.EpochInfos {
			if suggestedEpoch == 0 || int(info.GetLastCommittedEpoch()) < suggestedEpoch {
				suggestedEpoch = int(info.GetLastCommittedEpoch())
			}
		}
	}
	fmt.Printf("Last committed epoch: %d\n", suggestedEpoch)

	// Example 3: Get validator set
	fmt.Println("\n=== Getting Validator Set ===")
	validatorSet, err := relayClient.GetValidatorSet(ctx, nil)
	if err != nil {
		log.Printf("Failed to get validator set: %v", err)
	} else {
		fmt.Printf("Validator set version: %d\n", validatorSet.Version)
		fmt.Printf("Epoch: %d\n", validatorSet.Epoch)
		fmt.Printf("Status: %v\n", validatorSet.Status)
		fmt.Printf("Number of validators: %d\n", len(validatorSet.Validators))
		fmt.Printf("Quorum threshold: %s\n", validatorSet.QuorumThreshold)

		// Display some validator details
		if len(validatorSet.Validators) > 0 {
			firstValidator := validatorSet.Validators[0]
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

	fmt.Printf("Signature target id: %s\n", signResponse.GetSignatureTargetId())
	fmt.Printf("Epoch: %d\n", signResponse.Epoch)

	// Example 5: Get aggregation proof (this might fail if signing is not complete)
	fmt.Println("\n=== Getting Aggregation Proof ===")
	proofResponse, err := relayClient.GetAggregationProof(ctx, signResponse.GetSignatureTargetId())
	if err != nil {
		fmt.Printf("Could not get aggregation proof yet: %v\n", err)
	} else if proofResponse.AggregationProof != nil {
		proof := proofResponse.AggregationProof
		fmt.Printf("Proof length: %d bytes\n", len(proof.GetProof()))
		fmt.Printf("Message hash length: %d bytes\n", len(proof.GetMessageHash()))
	}

	// Example 6: Get individual signatures
	fmt.Println("\n=== Getting Individual Signatures ===")
	signaturesResponse, err := relayClient.GetSignatures(ctx, signResponse.GetSignatureTargetId())
	if err != nil {
		fmt.Printf("Could not get signatures yet: %v\n", err)
	} else {
		fmt.Printf("Number of signatures: %d\n", len(signaturesResponse.Signatures))

		for i, signature := range signaturesResponse.Signatures {
			fmt.Printf("Signature %d:\n", i+1)
			fmt.Printf("  - Signature length: %d bytes\n", len(signature.GetSignature()))
			fmt.Printf("  - Public key length: %d bytes\n", len(signature.GetPublicKey()))
			fmt.Printf("  - Message hash length: %d bytes\n", len(signature.GetMessageHash()))
		}
	}

	// Example 7: Sign and wait for completion (streaming)
	fmt.Println("\n=== Sign and Wait (Streaming) ===")
	messageToSignStream := []byte("Streaming example")

	fmt.Println("Starting streaming sign request... (ensure to run the script for all active relay servers)")

	stream, err := relayClient.SignMessageAndWait(ctx, keyTag, messageToSignStream, nil)
	if err != nil {
		log.Printf("Failed to start streaming sign: %v", err)
		return
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Stream completed")
				break
			}
			log.Printf("Stream error: %v", err)
			break
		}

		fmt.Printf("Status: %v\n", response.Status)
		fmt.Printf("Signature target id: %s\n", response.GetSignatureTargetId())
		fmt.Printf("Epoch: %d\n", response.Epoch)

		switch response.Status {
		case client.SigningStatus_SIGNING_STATUS_PENDING:
			fmt.Println("Request created, waiting for signatures...")
		case client.SigningStatus_SIGNING_STATUS_COMPLETED:
			fmt.Println("Signing completed!")
			if response.AggregationProof != nil {
				proof := response.AggregationProof
				fmt.Printf("Proof length: %d bytes\n", len(proof.GetProof()))
			}
			// Exit the streaming loop
			goto streamComplete
		case client.SigningStatus_SIGNING_STATUS_FAILED:
			fmt.Println("Signing failed")
			goto streamComplete
		case client.SigningStatus_SIGNING_STATUS_TIMEOUT:
			fmt.Println("Signing timed out")
			goto streamComplete
		case client.SigningStatus_SIGNING_STATUS_UNSPECIFIED:
			fmt.Println("Unknown Signing status : unspecified")
		default:
			fmt.Printf("Unknown status: %v\n", response.Status)
		}

		// Add a small delay to make the output more readable
		time.Sleep(100 * time.Millisecond)
	}

streamComplete:
	fmt.Println("\nExample completed")
}
