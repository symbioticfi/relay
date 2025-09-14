package signature_processor

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
)

func TestSignatureProcessor_ProcessSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		setupFunc              func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam
		expectSignatureRequest bool
		expectPendingExists    bool
		expectPendingRemoved   bool
		expectError            bool
		expectedErrorSubstring string
	}{
		{
			name: "new signature request - no quorum reached",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam {
				t.Helper()
				epoch := entity.Epoch(100)
				req := randomSignatureRequest(t, epoch)
				reqHash := req.Hash() // Use the actual request hash

				// Setup validator set header with high quorum threshold (1000)
				setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

				return entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      0,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)), // Not enough for quorum
					Epoch:            epoch,
					SignatureRequest: &req,
				}
			},
			expectSignatureRequest: true,
			expectPendingExists:    true,
			expectPendingRemoved:   false,
			expectError:            false,
		},
		{
			name: "new signature request - quorum reached",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam {
				t.Helper()
				epoch := entity.Epoch(101)
				req := randomSignatureRequest(t, epoch)
				reqHash := req.Hash() // Use the actual request hash

				// Setup validator set header with low quorum threshold (50)
				setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(50))

				return entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      0,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)), // Enough for quorum
					Epoch:            epoch,
					SignatureRequest: &req,
				}
			},
			expectSignatureRequest: true,
			expectPendingExists:    false, // Should be removed due to quorum
			expectPendingRemoved:   true,
			expectError:            false,
		},
		{
			name: "signature without signature request",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam {
				t.Helper()
				epoch := entity.Epoch(102)
				reqHash := randomHash(t)

				// Setup validator set header with high quorum threshold
				setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

				return entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      0,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)),
					Epoch:            epoch,
					SignatureRequest: nil, // No signature request
				}
			},
			expectSignatureRequest: false,
			expectPendingExists:    false,
			expectPendingRemoved:   false,
			expectError:            false,
		},
		{
			name: "multiple signatures - quorum reached on second",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam {
				t.Helper()
				epoch := entity.Epoch(103)
				req := randomSignatureRequest(t, epoch)
				reqHash := req.Hash() // Use the actual request hash

				// Setup validator set header with quorum threshold of 150
				setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(150))

				// First signature - not enough for quorum
				firstParam := entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      0,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)),
					Epoch:            epoch,
					SignatureRequest: &req,
				}

				processor, err := NewSignatureProcessor(Config{Repo: repo})
				require.NoError(t, err)

				err = processor.ProcessSignature(context.Background(), firstParam)
				require.NoError(t, err)

				// Verify pending exists after first signature
				_, err = repo.GetSignatureRequestsByEpochPending(context.Background(), epoch, 1, common.Hash{})
				require.NoError(t, err)

				// Return second signature that will reach quorum
				return entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      1,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)), // Total: 200 > 150
					Epoch:            epoch,
					SignatureRequest: nil, // Second signature doesn't include request again
				}
			},
			expectSignatureRequest: false,
			expectPendingExists:    false, // Should be removed after reaching quorum
			expectPendingRemoved:   true,
			expectError:            false,
		},
		{
			name: "missing validator set header",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SaveSignatureParam {
				t.Helper()
				epoch := entity.Epoch(999)
				reqHash := randomHash(t)

				// Don't setup validator set header - will cause error

				req := randomSignatureRequest(t, epoch)
				return entity.SaveSignatureParam{
					RequestHash:      reqHash,
					Key:              randomBytes(t, 48),
					Signature:        randomSignatureExtended(t),
					ActiveIndex:      0,
					VotingPower:      entity.ToVotingPower(big.NewInt(100)),
					Epoch:            epoch,
					SignatureRequest: &req,
				}
			},
			expectSignatureRequest: false,
			expectPendingExists:    false,
			expectPendingRemoved:   false,
			expectError:            true,
			expectedErrorSubstring: "failed to get active validator count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupTestRepository(t)
			param := tt.setupFunc(t, repo)

			processor, err := NewSignatureProcessor(Config{Repo: repo})
			require.NoError(t, err)

			err = processor.ProcessSignature(context.Background(), param)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedErrorSubstring != "" {
					require.Contains(t, err.Error(), tt.expectedErrorSubstring)
				}
				return
			}

			require.NoError(t, err)

			// Verify signature map was created/updated
			sigMap, err := repo.GetSignatureMap(context.Background(), param.RequestHash)
			require.NoError(t, err)
			require.Equal(t, param.RequestHash, sigMap.RequestHash)
			require.Equal(t, param.Epoch, sigMap.Epoch)
			require.True(t, sigMap.SignedValidatorsBitmap.Contains(param.ActiveIndex))

			// Verify signature was saved
			// Note: We can't easily test this without exposing GetSignature method

			// Verify signature request handling
			if tt.expectSignatureRequest && param.SignatureRequest != nil {
				// Should exist in main collection
				retrievedReq, err := repo.GetSignatureRequest(context.Background(), param.SignatureRequest.Hash())
				require.NoError(t, err)
				require.Equal(t, *param.SignatureRequest, retrievedReq)
			}

			// Verify pending collection state
			if tt.expectPendingExists {
				pendingReqs, err := repo.GetSignatureRequestsByEpochPending(context.Background(), param.Epoch, 10, common.Hash{})
				require.NoError(t, err)
				require.Len(t, pendingReqs, 1)
				if param.SignatureRequest != nil {
					require.Equal(t, *param.SignatureRequest, pendingReqs[0])
				}
			}

			if tt.expectPendingRemoved || !tt.expectPendingExists {
				pendingReqs, err := repo.GetSignatureRequestsByEpochPending(context.Background(), param.Epoch, 10, common.Hash{})
				require.NoError(t, err)
				require.Empty(t, pendingReqs)
			}
		})
	}
}

func TestSignatureProcessor_ProcessSignature_ConcurrentSignatures(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	epoch := entity.Epoch(200)
	req := randomSignatureRequest(t, epoch)
	reqHash := req.Hash() // Use the actual request hash

	// Setup validator set header with quorum threshold of 300
	setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(300))

	processor, err := NewSignatureProcessor(Config{Repo: repo})
	require.NoError(t, err)

	// Simulate 4 concurrent signatures
	signatures := []entity.SaveSignatureParam{
		{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: &req,
		},
		{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      1,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: nil,
		},
		{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      2,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: nil,
		},
		{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      3,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)), // Total: 400 > 300
			Epoch:            epoch,
			SignatureRequest: nil,
		},
	}

	// Process signatures sequentially (testing transaction consistency)
	for i, sig := range signatures {
		err := processor.ProcessSignature(context.Background(), sig)
		require.NoError(t, err, "Failed to process signature %d", i)
	}

	// Verify final state
	sigMap, err := repo.GetSignatureMap(context.Background(), reqHash)
	require.NoError(t, err)
	require.Equal(t, reqHash, sigMap.RequestHash)
	require.Equal(t, epoch, sigMap.Epoch)

	// All 4 validators should be present
	for i := uint32(0); i < 4; i++ {
		require.True(t, sigMap.SignedValidatorsBitmap.Contains(i), "Validator %d should be present", i)
	}

	// Pending collection should be empty (quorum reached)
	pendingReqs, err := repo.GetSignatureRequestsByEpochPending(context.Background(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingReqs)
}

func TestSignatureProcessor_ProcessSignature_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("duplicate signature for same validator", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(300)
		req := randomSignatureRequest(t, epoch)
		reqHash := req.Hash() // Use the actual request hash

		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

		param := entity.SaveSignatureParam{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: &req,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// First signature should succeed
		err = processor.ProcessSignature(context.Background(), param)
		require.NoError(t, err)

		// Duplicate signature should fail
		err = processor.ProcessSignature(context.Background(), param)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exist")
	})

	t.Run("zero voting power", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(301)

		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

		param := entity.SaveSignatureParam{
			RequestHash: randomHash(t), // Use random hash for this test since no SignatureRequest
			Key:         randomBytes(t, 48),
			Signature:   randomSignatureExtended(t),
			ActiveIndex: 0,
			VotingPower: entity.ToVotingPower(big.NewInt(0)), // Zero voting power
			Epoch:       epoch,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		err = processor.ProcessSignature(context.Background(), param)
		require.NoError(t, err)

		// Verify signature map still has zero voting power
		sigMap, err := repo.GetSignatureMap(context.Background(), param.RequestHash)
		require.NoError(t, err)
		require.Equal(t, "0", sigMap.CurrentVotingPower.String())
	})

	t.Run("exact quorum threshold", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(302)
		req := randomSignatureRequest(t, epoch)
		reqHash := req.Hash() // Use the actual request hash

		// Set quorum threshold to exactly 100
		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(100))

		param := entity.SaveSignatureParam{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)), // Exactly at threshold
			Epoch:            epoch,
			SignatureRequest: &req,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		err = processor.ProcessSignature(context.Background(), param)
		require.NoError(t, err)

		// Should reach quorum and remove from pending
		pendingReqs, err := repo.GetSignatureRequestsByEpochPending(context.Background(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingReqs)
	})
}

// Helper functions

func setupTestRepository(t *testing.T) *badger.Repository {
	t.Helper()
	repo, err := badger.New(badger.Config{Dir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := repo.Close()
		require.NoError(t, err)
	})
	return repo
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func randomHash(t *testing.T) common.Hash {
	t.Helper()
	return common.BytesToHash(randomBytes(t, 32))
}

func randomSignatureRequest(t *testing.T, epoch entity.Epoch) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}
}

func randomSignatureExtended(t *testing.T) entity.SignatureExtended {
	t.Helper()
	return entity.SignatureExtended{
		MessageHash: randomBytes(t, 32),
		Signature:   randomBytes(t, 64),
		PublicKey:   randomBytes(t, 48),
	}
}

func createValidatorSet(t *testing.T, epoch uint64, quorumThreshold *big.Int) entity.ValidatorSet {
	t.Helper()
	return createValidatorSetWithCount(t, epoch, quorumThreshold, 4) // Default to 4 validators for backward compatibility
}

func createValidatorSetWithCount(t *testing.T, epoch uint64, quorumThreshold *big.Int, validatorCount int) entity.ValidatorSet {
	t.Helper()

	validators := make([]entity.Validator, validatorCount)
	for i := 0; i < validatorCount; i++ {
		validators[i] = entity.Validator{
			Operator:    common.BytesToAddress(randomBytes(t, 20)),
			VotingPower: entity.ToVotingPower(big.NewInt(500)),
			IsActive:    true,
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: randomBytes(t, 32),
				},
			},
			Vaults: []entity.ValidatorVault{
				{
					ChainID:     1,
					Vault:       common.BytesToAddress(randomBytes(t, 20)),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
				},
			},
		}
	}

	validatorsList := entity.Validators(validators)
	validatorsList.SortByOperatorAddressAsc() // Sort validators by operator address

	return entity.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   entity.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  entity.ToVotingPower(quorumThreshold),
		Validators:       validatorsList,
		Status:           entity.HeaderCommitted,
	}
}

func setupValidatorSetHeader(t *testing.T, repo *badger.Repository, epoch uint64, quorumThreshold *big.Int) {
	t.Helper()
	vs := createValidatorSet(t, epoch, quorumThreshold)
	err := repo.SaveValidatorSet(t.Context(), vs)
	require.NoError(t, err)
}

func TestSignatureProcessor_ProcessAggregationProof(t *testing.T) {
	t.Parallel()

	t.Run("successfully processes aggregation proof", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(100)
		req := randomSignatureRequest(t, epoch)
		reqHash := req.Hash()

		// Create aggregated signature message
		msg := entity.AggregatedSignatureMessage{
			RequestHash: reqHash,
			KeyTag:      entity.KeyTag(1),
			Epoch:       epoch,
			AggregationProof: entity.AggregationProof{
				VerificationType: entity.VerificationTypeBlsBn254Simple,
				MessageHash:      randomBytes(t, 32),
				Proof:            randomBytes(t, 96),
			},
		}

		// Save pending aggregation proof first
		err := repo.SaveAggregationProofPending(t.Context(), reqHash, epoch)
		require.NoError(t, err)

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// Process aggregation proof
		err = processor.ProcessAggregationProof(t.Context(), msg)
		require.NoError(t, err)

		// Verify aggregation proof was saved
		savedProof, err := repo.GetAggregationProof(t.Context(), reqHash)
		require.NoError(t, err)
		require.Equal(t, msg.AggregationProof, savedProof)

		// Verify pending aggregation proof was removed
		pendingRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingRequests)
	})

	t.Run("handles missing pending aggregation proof gracefully", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(200)
		req := randomSignatureRequest(t, epoch)
		reqHash := req.Hash()

		msg := entity.AggregatedSignatureMessage{
			RequestHash: reqHash,
			KeyTag:      entity.KeyTag(1),
			Epoch:       epoch,
			AggregationProof: entity.AggregationProof{
				VerificationType: entity.VerificationTypeBlsBn254ZK,
				MessageHash:      randomBytes(t, 32),
				Proof:            randomBytes(t, 128),
			},
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// Process aggregation proof without pending entry (should succeed)
		err = processor.ProcessAggregationProof(t.Context(), msg)
		require.NoError(t, err)

		// Verify aggregation proof was still saved
		savedProof, err := repo.GetAggregationProof(t.Context(), reqHash)
		require.NoError(t, err)
		require.Equal(t, msg.AggregationProof, savedProof)
	})

	t.Run("fails when aggregation proof already exists", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(300)
		req := randomSignatureRequest(t, epoch)
		reqHash := req.Hash()

		msg := entity.AggregatedSignatureMessage{
			RequestHash: reqHash,
			KeyTag:      entity.KeyTag(1),
			Epoch:       epoch,
			AggregationProof: entity.AggregationProof{
				VerificationType: entity.VerificationTypeBlsBn254Simple,
				MessageHash:      randomBytes(t, 32),
				Proof:            randomBytes(t, 96),
			},
		}

		// Save aggregation proof first
		err := repo.SaveAggregationProof(t.Context(), reqHash, msg.AggregationProof)
		require.NoError(t, err)

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// Attempt to process same aggregation proof should fail
		err = processor.ProcessAggregationProof(t.Context(), msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to save aggregation proof")
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
	})
}

func TestSignatureProcessor_ProcessSignature_AggregationProofPending(t *testing.T) {
	t.Parallel()

	t.Run("saves aggregation proof pending for aggregation keys", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(400)
		req := randomSignatureRequest(t, epoch)
		req.KeyTag = entity.KeyTag(0) // Ensure it's an aggregation key (BlsBn254)
		reqHash := req.Hash()

		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

		param := entity.SaveSignatureParam{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: &req,
			KeyTag:           req.KeyTag,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// Process signature
		err = processor.ProcessSignature(t.Context(), param)
		require.NoError(t, err)

		// Verify signature request pending was saved
		pendingSignatureRequests, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingSignatureRequests, 1)
		require.Equal(t, reqHash, pendingSignatureRequests[0].Hash())

		// Verify aggregation proof pending was also saved
		pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingAggRequests, 1)
		require.Equal(t, reqHash, pendingAggRequests[0].Hash())
	})

	t.Run("does not save aggregation proof pending for non-aggregation keys", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(500)
		req := randomSignatureRequest(t, epoch)
		req.KeyTag = entity.KeyTag(0x10) // Ensure it's NOT an aggregation key (EcdsaSecp256k1)
		reqHash := req.Hash()

		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

		param := entity.SaveSignatureParam{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: &req,
			KeyTag:           req.KeyTag,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		// Process signature
		err = processor.ProcessSignature(t.Context(), param)
		require.NoError(t, err)

		// Verify signature request was saved but NOT to pending collection
		savedReq, err := repo.GetSignatureRequest(t.Context(), reqHash)
		require.NoError(t, err)
		require.Equal(t, req, savedReq)

		// Verify no pending signature requests (non-aggregation key)
		pendingSignatureRequests, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingSignatureRequests)

		// Verify no pending aggregation proof requests
		pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingAggRequests)
	})

	t.Run("integration test - full signature to aggregation proof flow", func(t *testing.T) {
		t.Parallel()

		repo := setupTestRepository(t)
		epoch := entity.Epoch(600)
		req := randomSignatureRequest(t, epoch)
		req.KeyTag = entity.KeyTag(0) // Aggregation key (BlsBn254)
		reqHash := req.Hash()

		setupValidatorSetHeader(t, repo, uint64(epoch), big.NewInt(1000))

		// Step 1: Process signature (should create pending aggregation proof)
		param := entity.SaveSignatureParam{
			RequestHash:      reqHash,
			Key:              randomBytes(t, 48),
			Signature:        randomSignatureExtended(t),
			ActiveIndex:      0,
			VotingPower:      entity.ToVotingPower(big.NewInt(100)),
			Epoch:            epoch,
			SignatureRequest: &req,
			KeyTag:           req.KeyTag,
		}

		processor, err := NewSignatureProcessor(Config{Repo: repo})
		require.NoError(t, err)

		err = processor.ProcessSignature(t.Context(), param)
		require.NoError(t, err)

		// Verify pending aggregation proof exists
		pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingAggRequests, 1)

		// Step 2: Process aggregation proof (should remove from pending)
		msg := entity.AggregatedSignatureMessage{
			RequestHash: reqHash,
			KeyTag:      req.KeyTag,
			Epoch:       epoch,
			AggregationProof: entity.AggregationProof{
				VerificationType: entity.VerificationTypeBlsBn254Simple,
				MessageHash:      randomBytes(t, 32),
				Proof:            randomBytes(t, 96),
			},
		}

		err = processor.ProcessAggregationProof(t.Context(), msg)
		require.NoError(t, err)

		// Verify aggregation proof was saved
		savedProof, err := repo.GetAggregationProof(t.Context(), reqHash)
		require.NoError(t, err)
		require.Equal(t, msg.AggregationProof, savedProof)

		// Verify pending aggregation proof was removed
		pendingAggRequests, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingAggRequests)
	})
}
