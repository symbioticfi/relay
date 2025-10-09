package entity_processor

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/errgroup"

	"github.com/symbioticfi/relay/internal/client/repository/badger"
	"github.com/symbioticfi/relay/pkg/signals"
	"github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"github.com/symbioticfi/relay/symbiotic/usecase/entity-processor/entity-processor/mocks"
)

func TestEntityProcessor_ProcessSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		setupFunc              func(t *testing.T, repo *badger.Repository) entity.SignatureExtended
		expectPendingExists    bool
		expectPendingRemoved   bool
		expectError            bool
		expectedErrorSubstring string
	}{
		{
			name: "new signature request - no quorum reached",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SignatureExtended {
				t.Helper()
				req := randomSignatureRequest(t, entity.Epoch(100))

				// Setup validator set header with high quorum threshold (1000)
				_, privateKeys := setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(1000))

				return signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req)
			},
			expectPendingExists:  false,
			expectPendingRemoved: false,
			expectError:          false,
		},
		{
			name: "new signature request - quorum reached",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SignatureExtended {
				t.Helper()
				epoch := entity.Epoch(101)
				req := randomSignatureRequest(t, epoch)

				// Setup validator set header with low quorum threshold (50)
				_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(50))

				return signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req)
			},
			expectPendingExists:  false, // Should be removed due to quorum
			expectPendingRemoved: true,
			expectError:          false,
		},
		{
			name: "signature without signature request",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SignatureExtended {
				t.Helper()
				epoch := entity.Epoch(102)

				// Setup validator set header with high quorum threshold
				_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(1000))

				return randomSignatureExtendedForKeyWithParams(t, privateKeys[0][15], entity.SignatureRequest{
					KeyTag:        entity.KeyTag(15),
					RequiredEpoch: epoch,
					Message:       nil,
				})
			},
			expectPendingExists:  false,
			expectPendingRemoved: false,
			expectError:          false,
		},
		{
			name: "multiple signatures - quorum reached on second",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SignatureExtended {
				t.Helper()
				epoch := entity.Epoch(103)
				req := randomSignatureRequest(t, epoch)

				// Setup validator set header with quorum threshold of 150
				_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(150))

				// First signature - not enough for quorum
				firstParam := signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req)

				processor, err := NewEntityProcessor(Config{
					Repo:                     repo,
					Aggregator:               createMockAggregator(t),
					AggProofSignal:           createMockAggProofSignal(t),
					SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
				})
				require.NoError(t, err)

				err = processor.ProcessSignature(t.Context(), firstParam)
				require.NoError(t, err)

				// Verify pending exists after first signature
				_, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 1, common.Hash{})
				require.NoError(t, err)

				// Return second signature that will reach quorum
				return signatureExtendedForRequest(t, privateKeys[1][req.KeyTag], req)
			},
			expectPendingExists:  false, // Should be removed after reaching quorum
			expectPendingRemoved: true,
			expectError:          false,
		},
		{
			name: "missing validator set header",
			setupFunc: func(t *testing.T, repo *badger.Repository) entity.SignatureExtended {
				t.Helper()
				// Don't setup validator set header - will cause error
				privateKey, err := crypto.GeneratePrivateKey(entity.KeyTypeBlsBn254)
				require.NoError(t, err)

				req := randomSignatureRequest(t, entity.Epoch(999))
				return randomSignatureExtendedForKeyWithParams(t, privateKey, req)
			},
			expectPendingExists:    false,
			expectPendingRemoved:   false,
			expectError:            true,
			expectedErrorSubstring: "validator not found for public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupTestRepository(t)
			param := tt.setupFunc(t, repo)

			processor, err := NewEntityProcessor(Config{
				Repo:                     repo,
				Aggregator:               createMockAggregator(t),
				AggProofSignal:           createMockAggProofSignal(t),
				SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
			})
			require.NoError(t, err)

			err = processor.ProcessSignature(t.Context(), param)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedErrorSubstring != "" {
					require.Contains(t, err.Error(), tt.expectedErrorSubstring)
				}
				return
			}

			require.NoError(t, err)

			// Verify signature map was created/updated
			sigMap, err := repo.GetSignatureMap(t.Context(), param.RequestID())
			require.NoError(t, err)
			require.Equal(t, param.RequestID(), sigMap.RequestID)
			require.Equal(t, param.Epoch, sigMap.Epoch)
			// Verify at least one validator is present in the bitmap
			require.Positive(t, sigMap.SignedValidatorsBitmap.GetCardinality(), "At least one validator should be present")

			// Verify pending collection state
			if tt.expectPendingExists {
				pendingReqs, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), param.Epoch, 10, common.Hash{})
				require.NoError(t, err)
				require.Len(t, pendingReqs, 1)
			}

			if tt.expectPendingRemoved || !tt.expectPendingExists {
				pendingReqs, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), param.Epoch, 10, common.Hash{})
				require.NoError(t, err)
				require.Empty(t, pendingReqs)
			}
		})
	}
}

func TestEntityProcessor_ProcessSignature_ConcurrentSignatures(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	epoch := entity.Epoch(200)
	req := randomSignatureRequest(t, epoch)

	// Setup validator set header with quorum threshold of 300
	_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(300))

	processor, err := NewEntityProcessor(Config{Repo: repo, Aggregator: createMockAggregator(t), AggProofSignal: createMockAggProofSignal(t), SignatureProcessedSignal: createMockSignatureProcessedSignal(t)})
	require.NoError(t, err)

	// Simulate 4 concurrent signatures
	signatures := []entity.SignatureExtended{
		signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req),
		signatureExtendedForRequest(t, privateKeys[1][req.KeyTag], req),
		signatureExtendedForRequest(t, privateKeys[2][req.KeyTag], req),
		signatureExtendedForRequest(t, privateKeys[3][req.KeyTag], req),
	}

	// Process signatures sequentially (testing transaction consistency)
	for i, sig := range signatures {
		err := processor.ProcessSignature(t.Context(), sig)
		require.NoError(t, err, "Failed to process signature %d", i)
	}

	// Verify final state
	sigMap, err := repo.GetSignatureMap(t.Context(), signatures[0].RequestID())
	require.NoError(t, err)
	require.Equal(t, signatures[0].RequestID(), sigMap.RequestID)
	require.Equal(t, epoch, sigMap.Epoch)

	// Since all signatures use the same key tag, they would resolve to the same validator
	// So we should have at least one validator present
	require.Positive(t, sigMap.SignedValidatorsBitmap.GetCardinality(), "At least one validator should be present")

	// Pending collection should be empty (quorum reached)
	pendingReqs, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingReqs)
}

func TestEntityProcessor_ProcessSignature_Conflict(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(200))

	vs, privateKeys := createValidatorSetWithCount(t, req.RequiredEpoch, big.NewInt(300), 20)
	err := repo.SaveValidatorSet(t.Context(), vs)
	require.NoError(t, err)

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	err = processor.ProcessSignature(t.Context(), signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req))
	require.NoError(t, err, "Failed to process signature")

	eg, egCtx := errgroup.WithContext(t.Context())
	for i := 1; i < len(privateKeys); i++ {
		eg.Go(func() error {
			return processor.ProcessSignature(egCtx, signatureExtendedForRequest(t, privateKeys[i][req.KeyTag], req))
		})
	}

	require.NoError(t, eg.Wait())
}

func TestEntityProcessor_ProcessSignature_DuplicateSignatureForSameValidator(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	epoch := entity.Epoch(300)
	req := randomSignatureRequest(t, epoch)

	_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(1000))

	param := signatureExtendedForRequest(t, privateKeys[0][15], req)

	processor, err := NewEntityProcessor(Config{Repo: repo, Aggregator: createMockAggregator(t), AggProofSignal: createMockAggProofSignal(t), SignatureProcessedSignal: createMockSignatureProcessedSignal(t)})
	require.NoError(t, err)

	// First signature should succeed
	err = processor.ProcessSignature(t.Context(), param)
	require.NoError(t, err)

	// Duplicate signature should fail
	err = processor.ProcessSignature(t.Context(), param)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exist")
}

func TestEntityProcessor_ProcessSignature_ExactQuorumThreshold(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	epoch := entity.Epoch(302)
	req := randomSignatureRequest(t, epoch)

	// Set quorum threshold to exactly 100
	_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(100))

	param := signatureExtendedForRequest(t, privateKeys[0][15], req)

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	err = processor.ProcessSignature(t.Context(), param)
	require.NoError(t, err)

	// Should reach quorum and remove from pending
	pendingReqs, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingReqs)
}

// Helper functions

func createMockAggregator(t *testing.T) *mocks.MockAggregator {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockAgg := mocks.NewMockAggregator(ctrl)
	// Default behavior: return true for verification
	mockAgg.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	return mockAgg
}

func createMockAggProofSignal(t *testing.T) *mocks.MockAggProofSignal {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockSignal := mocks.NewMockAggProofSignal(ctrl)
	// Default behavior: return nil for emit
	mockSignal.EXPECT().Emit(gomock.Any()).Return(nil).AnyTimes()
	return mockSignal
}

func createMockSignatureProcessedSignal(t *testing.T) *signals.Signal[entity.SignatureExtended] {
	t.Helper()
	return signals.New[entity.SignatureExtended](signals.DefaultConfig(), "test", nil)
}

func setupTestRepository(t *testing.T) *badger.Repository {
	t.Helper()
	repo, err := badger.New(badger.Config{Dir: t.TempDir(), Metrics: badger.DoNothingMetrics{}})
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

func randomSignatureRequest(t *testing.T, epoch entity.Epoch) entity.SignatureRequest {
	t.Helper()
	req := entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 512),
	}
	return req
}

func randomSignatureExtendedForKeyWithParams(t *testing.T, privateKey crypto.PrivateKey, req entity.SignatureRequest) entity.SignatureExtended {
	t.Helper()

	publicKey := privateKey.PublicKey()
	signature, messageHash, err := privateKey.Sign(req.Message)
	require.NoError(t, err)

	return entity.SignatureExtended{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		Signature:   signature,
		PublicKey:   publicKey.Raw(),
	}
}

// signatureExtendedForRequest creates a SignatureExtended for a given SignatureRequest using the same message
func signatureExtendedForRequest(t *testing.T, privateKey crypto.PrivateKey, req entity.SignatureRequest) entity.SignatureExtended {
	t.Helper()

	publicKey := privateKey.PublicKey()
	signature, messageHash, err := privateKey.Sign(req.Message)
	require.NoError(t, err)

	return entity.SignatureExtended{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		Signature:   signature,
		PublicKey:   publicKey.Raw(),
	}
}

func createValidatorSetWithCount(t *testing.T, epoch entity.Epoch, quorumThreshold *big.Int, validatorCount int) (entity.ValidatorSet, []map[entity.KeyTag]crypto.PrivateKey) {
	t.Helper()

	privateKeys := make([]map[entity.KeyTag]crypto.PrivateKey, validatorCount)
	validators := make([]entity.Validator, validatorCount)
	for i := 0; i < validatorCount; i++ {
		privateKeys[i] = make(map[entity.KeyTag]crypto.PrivateKey)
		// Generate a valid key for the validator
		privateKeyBLS, err := crypto.GeneratePrivateKey(entity.KeyTypeBlsBn254)
		require.NoError(t, err)
		privateKeys[i][15] = privateKeyBLS

		privateKeyECDSA, err := crypto.GeneratePrivateKey(entity.KeyTypeEcdsaSecp256k1)
		require.NoError(t, err)
		privateKeys[i][0x10] = privateKeyECDSA

		validators[i] = entity.Validator{
			Operator:    common.BytesToAddress(randomBytes(t, 20)),
			VotingPower: entity.ToVotingPower(big.NewInt(500)),
			IsActive:    true,
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: privateKeyBLS.PublicKey().OnChain(), // Use the actual on-chain representation
				},
				{
					Tag:     entity.KeyTag(0x10),
					Payload: privateKeyECDSA.PublicKey().OnChain(), // Use the actual on-chain representation
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
	}, privateKeys
}

func setupValidatorSetHeader(t *testing.T, repo *badger.Repository, epoch entity.Epoch, quorumThreshold *big.Int) (entity.ValidatorSet, []map[entity.KeyTag]crypto.PrivateKey) {
	t.Helper()
	vs, privateKeys := createValidatorSetWithCount(t, epoch, quorumThreshold, 4) // Default to 4 validators for backward compatibility
	err := repo.SaveValidatorSet(t.Context(), vs)
	require.NoError(t, err)
	return vs, privateKeys
}

func TestEntityProcessor_ProcessAggregationProof_SuccessfullyProcesses(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(100))

	msg := entity.AggregationProof{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: common.BytesToHash(randomBytes(t, 32)).Bytes(),
		Proof:       randomBytes(t, 96),
	}

	requestId := msg.RequestID()
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), requestId, req))

	// Setup validator set for this epoch (required by ProcessAggregationProof)
	setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(670))

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	// Process aggregation proof
	err = processor.ProcessAggregationProof(t.Context(), msg)
	require.NoError(t, err)

	// Verify aggregation proof was saved
	savedProof, err := repo.GetAggregationProof(t.Context(), msg.RequestID())
	require.NoError(t, err)
	require.Equal(t, msg, savedProof)

	// Verify pending aggregation proof was removed
	pendingRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), req.RequiredEpoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingRequests)
}

func TestEntityProcessor_ProcessAggregationProof_HandlesMissingPendingGracefully(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(200))
	msg := entity.AggregationProof{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: common.BytesToHash(randomBytes(t, 32)).Bytes(),
		Proof:       randomBytes(t, 128),
	}

	requestId := msg.RequestID()
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), requestId, req))

	// Setup validator set for this epoch (required by ProcessAggregationProof)
	setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(670))

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	// Process aggregation proof without pending entry (should succeed)
	err = processor.ProcessAggregationProof(t.Context(), msg)
	require.NoError(t, err)

	// Verify aggregation proof was still saved
	savedProof, err := repo.GetAggregationProof(t.Context(), msg.RequestID())
	require.NoError(t, err)
	require.Equal(t, msg, savedProof)
}

func TestEntityProcessor_ProcessAggregationProof_FailsWhenAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(300))
	msg := entity.AggregationProof{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: common.BytesToHash(randomBytes(t, 32)).Bytes(),
		Proof:       randomBytes(t, 96),
	}

	requestId := msg.RequestID()
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), requestId, req))

	// Setup validator set for this epoch (required by ProcessAggregationProof)
	setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(670))

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	err = processor.ProcessAggregationProof(t.Context(), msg)
	require.NoError(t, err)

	// Attempt to process same aggregation proof should fail
	err = processor.ProcessAggregationProof(t.Context(), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aggregation proof already exists for request ID")
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
}

func TestEntityProcessor_ProcessSignature_SavesAggregationProofPendingForAggregationKeys(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(400))

	_, privateKeys := setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(1000))
	param := randomSignatureExtendedForKeyWithParams(t, privateKeys[0][15], req)
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), param.RequestID(), req))

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	// Process signature
	err = processor.ProcessSignature(t.Context(), param)
	require.NoError(t, err)

	// Verify aggregation proof pending was also saved
	pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), req.RequiredEpoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Len(t, pendingAggRequests, 1)
}

func TestEntityProcessor_ProcessSignature_SaveAggregationProofPendingForNonAggregationKeys(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(500))
	req.KeyTag = entity.KeyTag(0x10) // Ensure it's NOT an aggregation key (EcdsaSecp256k1)

	_, privateKeys := setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(1000))

	param := randomSignatureExtendedForKeyWithParams(t, privateKeys[0][0x10], req)
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), param.RequestID(), req))

	// Verify signature request was saved but NOT to pending collection
	savedReq, err := repo.GetSignatureRequest(t.Context(), param.RequestID())
	require.NoError(t, err)
	require.Equal(t, req, savedReq)

	// Verify no pending aggregation proof requests
	pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), req.RequiredEpoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingAggRequests)
}

func TestEntityProcessor_ProcessSignature_FullSignatureToAggregationProofFlow(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)
	req := randomSignatureRequest(t, entity.Epoch(600))
	_, privateKeys := setupValidatorSetHeader(t, repo, req.RequiredEpoch, big.NewInt(1000))

	// Step 1: Process signature (should create pending aggregation proof)
	param := randomSignatureExtendedForKeyWithParams(t, privateKeys[0][15], req)
	require.NoError(t, repo.SaveSignatureRequest(t.Context(), param.RequestID(), req))

	processor, err := NewEntityProcessor(Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
	})
	require.NoError(t, err)

	err = processor.ProcessSignature(t.Context(), param)
	require.NoError(t, err)

	// Verify pending aggregation proof exists
	pendingAggRequests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), req.RequiredEpoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Len(t, pendingAggRequests, 1)

	// Step 2: Process aggregation proof (should remove from pending)
	msg := entity.AggregationProof{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: param.MessageHash,
		Proof:       randomBytes(t, 96),
	}

	err = processor.ProcessAggregationProof(t.Context(), msg)
	require.NoError(t, err)

	// Verify aggregation proof was saved
	savedProof, err := repo.GetAggregationProof(t.Context(), param.RequestID())
	require.NoError(t, err)
	require.Equal(t, msg, savedProof)

	// Verify pending aggregation proof was removed
	pendingAggRequests, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), req.RequiredEpoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, pendingAggRequests)
}
