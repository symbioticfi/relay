package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/core/entity"
	cryptoModule "github.com/symbioticfi/relay/core/usecase/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultECDSAKeyTag = 16
	secondaryBLSKeyTag = 11
)

// TestNonHeaderKeySignature tests signing with different non-header key types
func TestNonHeaderKeySignature(t *testing.T) {
	t.Log("Starting non-header key signature test...")

	deploymentData, err := loadDeploymentData()
	require.NoError(t, err, "Failed to load deployment data")

	endpoints := getRelayEndpoints(deploymentData.Env)
	expected := getExpectedDataFromContracts(t, deploymentData)

	msg := "random-stuff-test-" + rand.Text()

	testCases := []struct {
		name     string
		keyTag   entity.KeyTag
		testName string
	}{
		{
			name:     "ECDSA non-header key",
			keyTag:   entity.KeyTag(defaultECDSAKeyTag),
			testName: "Signing with ECDSA non-header key",
		},
		{
			name:     "BLS non-header key",
			keyTag:   entity.KeyTag(secondaryBLSKeyTag),
			testName: "Signing with BLS non-header key",
		},
	}

	t.Logf("Running signature test for string: %s", msg)
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			reqHash := ""
			for _, endpoint := range endpoints {
				func() {
					address := fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
					conn, err := grpc.NewClient(
						address,
						grpc.WithTransportCredentials(insecure.NewCredentials()),
					)
					require.NoErrorf(t, err, "Failed to connect to relay server at %s", address)
					defer conn.Close()

					client := apiv1.NewSymbioticClient(conn)
					var (
						resp *apiv1.SignMessageResponse
					)

					// retry sign call 3 times as it can get transaction conflict
					for attempts := 1; attempts <= 3; attempts++ {
						resp, err = client.SignMessage(context.Background(),
							&apiv1.SignMessageRequest{
								KeyTag:        uint32(tc.keyTag),
								Message:       []byte(msg),
								RequiredEpoch: nil,
							})
						if err == nil {
							break
						}
					}
					require.NoErrorf(t, err, "Failed to sign message with relay at %s", address)
					require.NotEmptyf(t, resp.RequestHash, "Empty request hash from relay at %s", address)
					if reqHash == "" {
						reqHash = resp.RequestHash
					} else {
						require.Equalf(t, reqHash, resp.RequestHash, "Mismatched request hash from relay at %s", address)
					}
				}()
			}

			// wait for signatures
			time.Sleep(5 * time.Second)

			t.Logf("Verifying signatures for request hash: %s", reqHash)

			timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			address := fmt.Sprintf("%s:%d", endpoints[0].Address, endpoints[0].Port)
			conn, err := grpc.NewClient(
				address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoErrorf(t, err, "Failed to connect to relay server at %s", address)
			defer conn.Close()

			client := apiv1.NewSymbioticClient(conn)

			for {
				select {
				case <-timeoutCtx.Done():
					t.Fatalf("Timed out waiting for all signatures for request hash: %s", reqHash)
				case <-ticker.C:
					resp, err := client.GetSignatures(context.Background(),
						&apiv1.GetSignaturesRequest{
							RequestHash: reqHash,
						})

					require.NoErrorf(t, err, "Failed to get signatures from relay at %s", address)

					if len(resp.GetSignatures()) != len(endpoints) {
						t.Logf("Received %d/%d signatures for request hash: %s. Waiting for all signatures...", len(resp.GetSignatures()), len(endpoints), reqHash)
						continue
					}
					t.Logf("All %d signatures received for request hash: %s", len(resp.GetSignatures()), reqHash)

					// verify signatures based on key type
					countMap := map[string]int{}
					for _, sig := range resp.GetSignatures() {
						found := false

						if tc.keyTag.Type() == entity.KeyTypeEcdsaSecp256k1 {
							// ECDSA signature verification using ethereum crypto
							publicKeyBytes, err := crypto.Ecrecover(sig.GetMessageHash(), sig.GetSignature())
							require.NoErrorf(t, err, "Failed to recover public key from signature for request hash: %s", reqHash)
							pubkey, err := crypto.UnmarshalPubkey(publicKeyBytes)
							require.NoErrorf(t, err, "Failed to unmarshal public key for request hash: %s", reqHash)
							addressBytes := crypto.PubkeyToAddress(*pubkey).Bytes()

						outer_ecdsa:
							for _, operator := range expected.ValidatorSet.Validators {
								for _, key := range operator.Keys {
									if key.Tag != tc.keyTag {
										continue
									}
									// the contract stores 32 bytes padded address for ecdsa addrs,
									// so stripping first 12 bytes to get to the address
									if bytes.Equal(key.Payload[12:], addressBytes) {
										countMap[operator.Operator.String()]++
										found = true
										break outer_ecdsa
									}
								}
							}
						} else if tc.keyTag.Type() == entity.KeyTypeBlsBn254 {
							// Create public key from stored payload
							publicKey, err := cryptoModule.NewPublicKey(tc.keyTag.Type(), sig.GetPublicKey())
							require.NoErrorf(t, err, "Failed to create public key for request hash: %s", reqHash)

						outer_bls:
							for _, operator := range expected.ValidatorSet.Validators {
								for _, key := range operator.Keys {
									if key.Tag != tc.keyTag {
										continue
									}
									if !bytes.Equal(publicKey.OnChain(), key.Payload) {
										continue
									}

									// Verify signature using BLS verification
									err = publicKey.VerifyWithHash(sig.GetMessageHash(), sig.GetSignature())
									if err == nil {
										countMap[operator.Operator.String()]++
										found = true
										break outer_bls
									}
								}
							}
						}

						require.Truef(t, found, "Signature verification failed for key type %v for request hash: %s", tc.keyTag.Type(), reqHash)
					}

					// check for proof
					proof, err := client.GetAggregationProof(context.Background(), &apiv1.GetAggregationProofRequest{
						RequestHash: reqHash,
					})
					if tc.keyTag.Type() == entity.KeyTypeEcdsaSecp256k1 {
						require.Errorf(t, err, "Expected no aggregation proof for ECDSA key type for request hash: %s", reqHash)
					} else if tc.keyTag.Type() == entity.KeyTypeBlsBn254 {
						require.NoErrorf(t, err, "Failed to get aggregation proof for BLS key type for request hash: %s", reqHash)
						require.NotNilf(t, proof, "Expected aggregation proof for BLS key type for request hash: %s", reqHash)
						require.NotEmptyf(t, proof.GetAggregationProof().GetProof(), "Empty aggregation proof for BLS key type for request hash: %s", reqHash)
					}
					require.Lenf(t, countMap, len(expected.ValidatorSet.Validators), "Number of unique valid signatures does not match number of validators for request hash: %s", reqHash)
					t.Logf("%s test completed successfully", tc.name)
					return
				}
			}
		})
	}
}
