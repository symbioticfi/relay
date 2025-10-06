package main

//
//import (
//	"context"
//	"crypto/rand"
//	"fmt"
//	"log"
//	"math/big"
//	"runtime"
//	"sync"
//	"time"
//
//	"github.com/go-errors/errors"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
//
//	client "github.com/symbioticfi/relay/api/client/v1"
//)
//
//// RelayClient wraps the Symbiotic client with helpful methods
//type RelayClient struct {
//	client *client.SymbioticClient
//	conn   *grpc.ClientConn
//}
//
//// NewRelayClient creates a new client connected to the specified server URL
//func NewRelayClient(serverURL string) (*RelayClient, error) {
//	// Create gRPC connection
//	conn, err := grpc.NewClient(serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		return nil, errors.Errorf("failed to connect to server: %w", err)
//	}
//
//	// Create the symbiotic client
//	symbioticClient := client.NewSymbioticClient(conn)
//
//	fmt.Printf("Connected to Symbiotic Relay at %s\n", serverURL)
//
//	return &RelayClient{
//		client: symbioticClient,
//		conn:   conn,
//	}, nil
//}
//
//// Close closes the gRPC connection
//func (rc *RelayClient) Close() error {
//	return rc.conn.Close()
//}
//
//// GetCurrentEpoch gets the current epoch information
//func (rc *RelayClient) GetCurrentEpoch(ctx context.Context) (*client.GetCurrentEpochResponse, error) {
//	req := &client.GetCurrentEpochRequest{}
//	return rc.client.GetCurrentEpoch(ctx, req)
//}
//
//// GetLastAllCommitted gets the last all committed epochs for all chains
//func (rc *RelayClient) GetLastAllCommitted(ctx context.Context) (*client.GetLastAllCommittedResponse, error) {
//	req := &client.GetLastAllCommittedRequest{}
//	return rc.client.GetLastAllCommitted(ctx, req)
//}
//
//// SignMessage signs a message using the specified key tag
//func (rc *RelayClient) SignMessage(ctx context.Context, keyTag uint32, message []byte, requiredEpoch *uint64) (*client.SignMessageResponse, error) {
//	req := &client.SignMessageRequest{
//		KeyTag:        keyTag,
//		Message:       message,
//		RequiredEpoch: requiredEpoch,
//	}
//	return rc.client.SignMessage(ctx, req)
//}
//
//// GetAggregationProof gets aggregation proof for a specific request
//func (rc *RelayClient) GetAggregationProof(ctx context.Context, requestID string) (*client.GetAggregationProofResponse, error) {
//	req := &client.GetAggregationProofRequest{
//		RequestId: requestID,
//	}
//	return rc.client.GetAggregationProof(ctx, req)
//}
//
//// GetSignatures gets individual signatures for a request
//func (rc *RelayClient) GetSignatures(ctx context.Context, requestID string) (*client.GetSignaturesResponse, error) {
//	req := &client.GetSignaturesRequest{
//		RequestId: requestID,
//	}
//	return rc.client.GetSignatures(ctx, req)
//}
//
//// GetValidatorSet gets validator set information
//func (rc *RelayClient) GetValidatorSet(ctx context.Context, epoch *uint64) (*client.GetValidatorSetResponse, error) {
//	req := &client.GetValidatorSetRequest{
//		Epoch: epoch,
//	}
//	return rc.client.GetValidatorSet(ctx, req)
//}
//
//// SignMessageAndWait signs a message and waits for aggregation via streaming response
//func (rc *RelayClient) SignMessageAndWait(ctx context.Context, keyTag uint32, message []byte, requiredEpoch *uint64) (grpc.ServerStreamingClient[client.SignMessageWaitResponse], error) {
//	req := &client.SignMessageWaitRequest{
//		KeyTag:        keyTag,
//		Message:       message,
//		RequiredEpoch: requiredEpoch,
//	}
//	return rc.client.SignMessageWait(ctx, req)
//}
//
//func main() {
//
//	runtime.GOMAXPROCS(40)
//
//	startPort := 8081
//	numberOfRelays := 4
//
//	var relayClients []*RelayClient
//
//	for i := 0; i < numberOfRelays; i++ {
//		serverURL := fmt.Sprintf("localhost:%d", startPort+i)
//		relayClient, err := NewRelayClient(serverURL)
//		if err != nil {
//			log.Fatalf("failed to create relay client: %v", err)
//		}
//		_, err = relayClient.GetCurrentEpoch(context.Background())
//		if err != nil {
//			log.Fatalf("failed to perform test request: %v", err)
//		}
//		relayClients = append(relayClients, relayClient)
//	}
//
//	requestedMessages := 0
//	requestedSignatures := 0
//	receivedRequestIds := 0
//	receivedProofs := make(map[string]bool)
//	receivedProofsMutex := sync.Mutex{}
//
//	globalStart := time.Now()
//	var wg sync.WaitGroup
//
//	for k := 0; k < 100; k++ {
//		start := time.Now()
//		epoch, err := relayClients[0].GetCurrentEpoch(context.Background())
//		if err != nil {
//			log.Fatalf("failed to get current epoch: %v", err)
//		}
//		log.Printf("Current epoch: %d, request duration %f\n", epoch.Epoch, time.Since(start).Seconds())
//
//		random, _ := rand.Int(rand.Reader, big.NewInt(10000000000000))
//		messageToSign := []byte("Hello, Symbiotic!" + random.String())
//		requestedMessages += 1
//
//		wg.Add(numberOfRelays)
//		// Example 1: Get current epoch
//		for i, relayClient := range relayClients {
//			go func() {
//				defer wg.Done()
//
//				start := time.Now()
//				log.Printf("Sending message to Relay node %d at epoch %d\n", i, epoch.Epoch)
//				requestedSignatures += 1
//				resp, err := relayClient.SignMessage(context.Background(), uint32(15), messageToSign, &epoch.Epoch)
//				if err != nil {
//					log.Printf("Failed to sign, node %d: %v, took: %f seconds", i, err, time.Since(start).Seconds())
//					return
//				}
//				log.Printf("Successfully send sign message to node %d, request id: %s request duration: %f", i, resp.RequestId, time.Since(start).Seconds())
//				receivedRequestIds += 1
//
//				requestId := resp.RequestId
//				for i := 0; i < 500; i++ {
//					resp, err := relayClient.GetAggregationProof(context.Background(), requestId)
//					if err == nil {
//						log.Printf("Successfully got aggregation proof from Relay node %d proof length %d in %f seconds\n", i, len(resp.GetAggregationProof().Proof), time.Since(start).Seconds())
//						receivedProofsMutex.Lock()
//						receivedProofs[requestId] = true
//						receivedProofsMutex.Unlock()
//						return
//					}
//					time.Sleep(5 * time.Millisecond)
//				}
//				log.Printf("Failed to get aggregation proof from Relay node %d in %f seconds\n", i, time.Since(start).Seconds())
//			}()
//
//		}
//	}
//	wg.Wait()
//
//	log.Printf("Reqeusted messages: %d", requestedMessages)
//	log.Printf("Received proofs: %d", len(receivedProofs))
//	log.Printf("Requested signatures: %d", requestedSignatures)
//	log.Printf("Received request ids: %d", receivedRequestIds)
//	log.Printf("Took %f seconds\n", time.Since(globalStart).Seconds())
//	log.Printf("Proof fail ratio: %f\n", (float64(requestedMessages-len(receivedProofs)))/float64(requestedMessages))
//	log.Printf("AVG proof/sec rate: %f\n", float64(len(receivedProofs))/time.Since(globalStart).Seconds())
//}
