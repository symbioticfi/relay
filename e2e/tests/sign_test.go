package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/core/entity"
	cryptoModule "github.com/symbioticfi/relay/core/usecase/crypto"
)

const (
	defaultECDSAKeyTag = 16
	secondaryBLSKeyTag = 11
)

// TestNonHeaderKeySignature tests signing with different non-header key types
func TestNonHeaderKeySignature(t *testing.T) {
	t.Log("Starting non-header key signature test...")

	deploymentData, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

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
			signatureTargetID := ""
			for i := range globalTestEnv.Containers {
				func() {
					client := globalTestEnv.GetGRPCClient(t, i)

					var resp *apiv1.SignMessageResponse
					// retry sign call 3 times as it can get transaction conflict
					for attempts := 1; attempts <= 3; attempts++ {
						resp, err = client.SignMessage(t.Context(),
							&apiv1.SignMessageRequest{
								KeyTag:        uint32(tc.keyTag),
								Message:       []byte(msg),
								RequiredEpoch: nil,
							})
						if err == nil {
							break
						}
					}
					require.NoErrorf(t, err, "Failed to sign message with relay at %d", i)
					require.NotEmptyf(t, resp.SignatureTargetId, "Empty signature target id from relay at %d", i)
					if signatureTargetID == "" {
						signatureTargetID = resp.GetSignatureTargetId()
					} else {
						require.Equalf(t, signatureTargetID, resp.SignatureTargetId, "Mismatched signature target id from relay at %d", i)
					}
				}()
			}

			// wait for signatures
			time.Sleep(5 * time.Second)

			t.Logf("Verifying signatures for signature target id: %s", signatureTargetID)

			timeoutCtx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			client := globalTestEnv.GetGRPCClient(t, 0)

			for {
				select {
				case <-timeoutCtx.Done():
					t.Fatalf("Timed out waiting for all signatures for signature target id: %s", signatureTargetID)
				case <-ticker.C:
					resp, err := client.GetSignatures(t.Context(),
						&apiv1.GetSignaturesRequest{
							SignatureTargetId: signatureTargetID,
						})

					require.NoErrorf(t, err, "Failed to get signatures from relay at %d", 0)

					if tc.keyTag.Type() == entity.KeyTypeEcdsaSecp256k1 && len(resp.GetSignatures()) != len(globalTestEnv.Containers) {
						// expect all n signatures for ECDSA
						t.Logf("Received %d/%d signatures for signature target id: %s. Waiting for all signatures...", len(resp.GetSignatures()), len(globalTestEnv.Containers), signatureTargetID)
						continue
					} else if tc.keyTag.Type() == entity.KeyTypeBlsBn254 && (len(globalTestEnv.Containers)*2/3+1) > len(resp.GetSignatures()) {
						// need at least 2/3 signatures for BLS, signers skip signing is proof is already generated so we may not get all n sigs
						t.Logf("Received %d/%d signatures for signature target id: %s. Waiting for all signatures...", len(resp.GetSignatures()), len(globalTestEnv.Containers), signatureTargetID)
						continue
					}
					t.Logf("All %d signatures received for signature target id: %s", len(resp.GetSignatures()), signatureTargetID)

					// verify signatures based on key type
					countMap := map[string]int{}
					for _, sig := range resp.GetSignatures() {
						found := false

						if tc.keyTag.Type() == entity.KeyTypeEcdsaSecp256k1 {
							// ECDSA signature verification using ethereum crypto
							publicKeyBytes, err := crypto.Ecrecover(sig.GetMessageHash(), sig.GetSignature())
							require.NoErrorf(t, err, "Failed to recover public key from signature for signature target id: %s", signatureTargetID)
							pubkey, err := crypto.UnmarshalPubkey(publicKeyBytes)
							require.NoErrorf(t, err, "Failed to unmarshal public key for signature target id: %s", signatureTargetID)
							addressBytes := crypto.PubkeyToAddress(*pubkey).Bytes()

						outerECDSA:
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
										break outerECDSA
									}
								}
							}
						} else if tc.keyTag.Type() == entity.KeyTypeBlsBn254 {
							// Create public key from stored payload
							publicKey, err := cryptoModule.NewPublicKey(tc.keyTag.Type(), sig.GetPublicKey())
							require.NoErrorf(t, err, "Failed to create public key for signature target id: %s", signatureTargetID)

						outerBLS:
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
										break outerBLS
									}
								}
							}
						}

						require.Truef(t, found, "Signature verification failed for key type %v for signature target id: %s", tc.keyTag.Type(), signatureTargetID)
					}

					// check for proof
					proof, err := client.GetAggregationProof(t.Context(), &apiv1.GetAggregationProofRequest{
						SignatureTargetId: signatureTargetID,
					})
					if tc.keyTag.Type() == entity.KeyTypeEcdsaSecp256k1 {
						require.Errorf(t, err, "Expected no aggregation proof for ECDSA key type for signature target id: %s", signatureTargetID)
					} else if tc.keyTag.Type() == entity.KeyTypeBlsBn254 {
						require.NoErrorf(t, err, "Failed to get aggregation proof for BLS key type for signature target id: %s", signatureTargetID)
						require.NotNilf(t, proof, "Expected aggregation proof for BLS key type for signature target id: %s", signatureTargetID)
						require.NotEmptyf(t, proof.GetAggregationProof().GetProof(), "Empty aggregation proof for BLS key type for signature target id: %s", signatureTargetID)
					}
					require.Lenf(t, countMap, len(resp.GetSignatures()), "Number of unique valid signatures does not match number of validators for signature target id: %s", signatureTargetID)
					t.Logf("%s test completed successfully", tc.name)
					return
				}
			}
		})
	}
}
