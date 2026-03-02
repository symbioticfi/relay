package sync_provider

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/client/repository/badger"
	"github.com/symbioticfi/relay/internal/entity"
	entity_processor "github.com/symbioticfi/relay/internal/usecase/entity-processor"
	"github.com/symbioticfi/relay/internal/usecase/entity-processor/mocks"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestAskSignatures_HandleWantSignaturesRequest_Integration(t *testing.T) {
	requesterRepo := createTestRepo(t)
	defer requesterRepo.Close()
	peerRepo := createTestRepo(t)
	defer peerRepo.Close()

	// Create test data
	privateKey := newPrivateKey(t)
	privateKey1 := newPrivateKey(t)
	signatureRequest := createTestSignatureRequest(t)
	validatorSet := createTestValidatorSet(t, privateKey, privateKey1)
	nextValsetData := entity.NextValsetData{
		NextValidatorSet:     validatorSet,
		NextNetworkConfig:    randomNetworkConfig(),
		PrevValidatorSet:     validatorSet,
		PrevNetworkConfig:    randomNetworkConfig(),
		SignatureRequest:     &signatureRequest,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{},
	}
	require.NoError(t, peerRepo.SaveNextValsetData(t.Context(), nextValsetData))
	require.NoError(t, requesterRepo.SaveNextValsetData(t.Context(), nextValsetData))

	peerEntityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     peerRepo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)

	requesterEntityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     requesterRepo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)

	signature, hash, err := privateKey.Sign(signatureRequest.Message)
	require.NoError(t, err)

	// Save signature request and signature on peer
	param := symbiotic.Signature{
		MessageHash: hash,
		Signature:   signature,
		PublicKey:   privateKey.PublicKey(),
		Epoch:       signatureRequest.RequiredEpoch,
		KeyTag:      signatureRequest.KeyTag,
	}
	require.NoError(t, peerEntityProcessor.ProcessSignature(t.Context(), param, false))

	requestID := param.RequestID()

	signature1, _, err := privateKey1.Sign(signatureRequest.Message)
	require.NoError(t, err)
	param1 := symbiotic.Signature{
		MessageHash: hash,
		Signature:   signature1,
		PublicKey:   privateKey1.PublicKey(),
		Epoch:       signatureRequest.RequiredEpoch,
		KeyTag:      signatureRequest.KeyTag,
	}
	require.NoError(t, requesterEntityProcessor.ProcessSignature(t.Context(), param1, false))

	// Save signature request on both repos so peer can respond to requests
	require.NoError(t, peerRepo.SaveSignatureRequest(t.Context(), requestID, signatureRequest))
	require.NoError(t, requesterRepo.SaveSignatureRequest(t.Context(), requestID, signatureRequest))

	// Requester needs SignatureMap for BuildWantSignaturesRequest to work
	signatureMap := entity.NewSignatureMap(requestID, signatureRequest.RequiredEpoch, uint32(len(validatorSet.Validators)))
	require.NoError(t, requesterRepo.UpdateSignatureMap(t.Context(), signatureMap))

	// Create peer syncer first (with a temporary mock)
	peerSyncer, err := New(Config{
		Repo:                        peerRepo,
		EntityProcessor:             peerEntityProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
		MaxAggProofRequestsPerSync:  100,
		MaxResponseAggProofCount:    100,
	})
	require.NoError(t, err)

	// Create requester syncer
	requesterSyncer, err := New(Config{
		Repo:                        requesterRepo,
		EntityProcessor:             requesterEntityProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
		MaxAggProofRequestsPerSync:  100,
		MaxResponseAggProofCount:    100,
	})
	require.NoError(t, err)

	// Verify requester initially has no signatures
	initialSignatures, err := requesterRepo.GetAllSignatures(t.Context(), requestID)
	require.NoError(t, err)
	require.Len(t, initialSignatures, 1) // Already has one signature from param1

	// Verify requester has signature request
	_, err = requesterRepo.GetSignatureRequest(t.Context(), requestID)
	require.NoError(t, err)

	// Call BuildWantSignaturesRequest on requester
	request, err := requesterSyncer.BuildWantSignaturesRequest(t.Context())
	require.NoError(t, err)

	response, err := peerSyncer.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, response.Signatures, 1)

	stat := requesterSyncer.ProcessReceivedSignatures(t.Context(), response, request.WantSignatures)
	require.Equal(t, 0, stat.TotalErrors())

	// Verify requester now has the signature
	finalSignatures, err := requesterRepo.GetAllSignatures(t.Context(), requestID)
	require.NoError(t, err)
	require.Len(t, finalSignatures, 2)

	// Verify the signature is correct
	require.Equal(t, privateKey.PublicKey(), finalSignatures[0].PublicKey)
	require.NoError(t, privateKey.PublicKey().Verify(signatureRequest.Message, finalSignatures[0].Signature))
	require.Equal(t, privateKey1.PublicKey(), finalSignatures[1].PublicKey)
	require.NoError(t, privateKey1.PublicKey().Verify(signatureRequest.Message, finalSignatures[1].Signature))
}

func createMockAggregator(t *testing.T) *mocks.MockAggregator {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockAgg := mocks.NewMockAggregator(ctrl)
	// Default behavior: return true for verification
	mockAgg.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
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

func createMockSignatureProcessedSignal(t *testing.T) *signals.Signal[symbiotic.Signature] {
	t.Helper()
	return signals.New[symbiotic.Signature](signals.DefaultConfig(), "test", nil)
}

func createTestRepo(t *testing.T) *badger.Repository {
	t.Helper()
	repo, err := badger.New(badger.Config{
		Dir:            t.TempDir(),
		Metrics:        badger.DoNothingMetrics{},
		BlockCacheSize: -1,
	})
	require.NoError(t, err)
	return repo
}

func createTestSignatureRequest(t *testing.T) symbiotic.SignatureRequest {
	t.Helper()
	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: symbiotic.Epoch(1),
		Message:       randomBytes(t, 100), // Random message makes each request unique
	}
}

func newPrivateKey(t *testing.T) crypto.PrivateKey {
	t.Helper()
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	require.NoError(t, err)

	privateKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, privateKeyBytes)
	require.NoError(t, err)
	return privateKey
}

func createTestValidatorSet(t *testing.T, privateKey ...crypto.PrivateKey) symbiotic.ValidatorSet {
	t.Helper()
	validators := make([]symbiotic.Validator, len(privateKey))
	for i, pk := range privateKey {
		validators[i] = symbiotic.Validator{
			Operator:    common.HexToAddress(fmt.Sprintf("0x%d", i+1)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(15),
					Payload: pk.PublicKey().OnChain(),
				},
			},
		}
	}
	return symbiotic.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  symbiotic.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(670)),
		Validators:      validators,
	}
}

func createTestValidatorSetWithMultipleValidators(t *testing.T, count int) (symbiotic.ValidatorSet, []crypto.PrivateKey) {
	t.Helper()
	privateKeys := make([]crypto.PrivateKey, count)

	validators := make([]symbiotic.Validator, count)
	for i := 0; i < count; i++ {
		privateKey := newPrivateKey(t)
		privateKeys[i] = privateKey

		validators[i] = symbiotic.Validator{
			Operator:    common.HexToAddress(fmt.Sprintf("0x%d", i+1)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(15),
					Payload: privateKey.PublicKey().OnChain(), // Same key for all validators for simplicity
				},
			},
		}
	}

	return symbiotic.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  symbiotic.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(670)),
		Validators:      validators,
	}, privateKeys
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func TestHandleWantSignaturesRequest_EmptyRequest(t *testing.T) {
	t.Parallel()

	repo := createTestRepo(t)
	defer repo.Close()

	entityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)

	syncer, err := New(Config{
		Repo:                        repo,
		EntityProcessor:             entityProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
		MaxAggProofRequestsPerSync:  100,
		MaxResponseAggProofCount:    100,
	})
	require.NoError(t, err)

	t.Run("completely empty request", func(t *testing.T) {
		request := entity.WantSignaturesRequest{
			WantSignatures: map[common.Hash]entity.Bitmap{},
		}

		response, err := syncer.HandleWantSignaturesRequest(t.Context(), request)
		require.NoError(t, err)
		require.Empty(t, response.Signatures)
	})
}

func TestHandleWantSignaturesRequest_MaxResponseSignatureCountLimit(t *testing.T) {
	t.Parallel()

	repo := createTestRepo(t)
	defer repo.Close()

	// Create test data with multiple signatures
	validatorSet, privateKeys := createTestValidatorSetWithMultipleValidators(t, 5) // Create 5 validators
	signatureRequest := createTestSignatureRequest(t)

	nextValsetData := entity.NextValsetData{
		NextValidatorSet:     validatorSet,
		NextNetworkConfig:    randomNetworkConfig(),
		PrevValidatorSet:     validatorSet,
		PrevNetworkConfig:    randomNetworkConfig(),
		SignatureRequest:     &signatureRequest,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{},
	}
	// Setup repository with validator set and signature request
	require.NoError(t, repo.SaveNextValsetData(t.Context(), nextValsetData))

	// Store multiple signatures by validator index
	entityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)
	var requestID common.Hash
	// Save signatures for multiple validator indices (simulate multiple validators)
	for i := uint32(0); i < 5; i++ {
		signature, hash, err := privateKeys[i].Sign(signatureRequest.Message)
		require.NoError(t, err)

		param := symbiotic.Signature{
			MessageHash: hash,
			KeyTag:      signatureRequest.KeyTag,
			Epoch:       signatureRequest.RequiredEpoch,
			PublicKey:   privateKeys[i].PublicKey(),
			Signature:   signature,
		}
		if i == 0 {
			requestID = param.RequestID()
			require.NoError(t, repo.SaveSignatureRequest(t.Context(), requestID, signatureRequest))
		}
		require.NoError(t, entityProcessor.ProcessSignature(t.Context(), param, false))
	}

	t.Run("limit exceeded with single request", func(t *testing.T) {
		syncer, err := New(Config{
			Repo:                        repo,
			EntityProcessor:             entityProcessor,
			EpochsToSync:                1,
			MaxSignatureRequestsPerSync: 100,
			MaxResponseSignatureCount:   2, // Low limit
			MaxAggProofRequestsPerSync:  100,
			MaxResponseAggProofCount:    100,
		})
		require.NoError(t, err)

		request := entity.WantSignaturesRequest{
			WantSignatures: map[common.Hash]entity.Bitmap{
				requestID: entity.NewBitmapOf(0, 1, 2, 3, 4), // Request all 5 signatures
			},
		}

		response, err := syncer.HandleWantSignaturesRequest(t.Context(), request)
		require.NoError(t, err)
		require.Len(t, response.Signatures, 1)
		require.Len(t, response.Signatures[requestID], 2) // Should return only 2 signatures due to limit
	})

	t.Run("limit respected", func(t *testing.T) {
		syncer, err := New(Config{
			Repo:                        repo,
			EntityProcessor:             entityProcessor,
			EpochsToSync:                1,
			MaxSignatureRequestsPerSync: 100,
			MaxResponseSignatureCount:   3, // Allow 3 signatures
			MaxAggProofRequestsPerSync:  100,
			MaxResponseAggProofCount:    100,
		})
		require.NoError(t, err)

		request := entity.WantSignaturesRequest{
			WantSignatures: map[common.Hash]entity.Bitmap{
				requestID: entity.NewBitmapOf(0, 1, 2), // Request exactly 3 signatures
			},
		}

		response, err := syncer.HandleWantSignaturesRequest(t.Context(), request)
		require.NoError(t, err)
		require.Len(t, response.Signatures, 1)
		require.Len(t, response.Signatures[requestID], 3)
	})
}

func TestHandleWantSignaturesRequest_MultipleRequestIDs(t *testing.T) {
	t.Parallel()

	repo := createTestRepo(t)
	defer repo.Close()

	// Store signatures for both requests
	entityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)

	syncer, err := New(Config{
		Repo:                        repo,
		EntityProcessor:             entityProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
		MaxAggProofRequestsPerSync:  100,
		MaxResponseAggProofCount:    100,
	})
	require.NoError(t, err)

	// Create test data
	validatorSet, privateKeys := createTestValidatorSetWithMultipleValidators(t, 2) // Create 2 validators
	signatureRequest1 := createTestSignatureRequest(t)
	signatureRequest2 := createTestSignatureRequest(t)

	nextValsetData := entity.NextValsetData{
		NextValidatorSet:     validatorSet,
		NextNetworkConfig:    randomNetworkConfig(),
		PrevValidatorSet:     validatorSet,
		PrevNetworkConfig:    randomNetworkConfig(),
		SignatureRequest:     &signatureRequest1,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{},
	}

	// Setup repository
	require.NoError(t, repo.SaveNextValsetData(t.Context(), nextValsetData))

	signature, hash, err := privateKeys[0].Sign(signatureRequest1.Message)
	require.NoError(t, err)

	// Save signature for first request
	param1 := symbiotic.Signature{
		MessageHash: hash,
		Epoch:       signatureRequest1.RequiredEpoch,
		KeyTag:      signatureRequest1.KeyTag,
		Signature:   signature,
		PublicKey:   privateKeys[0].PublicKey(),
	}

	require.NoError(t, repo.SaveSignatureRequest(t.Context(), param1.RequestID(), signatureRequest1))
	require.NoError(t, entityProcessor.ProcessSignature(t.Context(), param1, false))

	// Save signature for second request
	signature2, hash2, err := privateKeys[1].Sign(signatureRequest2.Message)
	require.NoError(t, err)

	param2 := symbiotic.Signature{
		MessageHash: hash2,
		Epoch:       signatureRequest2.RequiredEpoch,
		KeyTag:      signatureRequest2.KeyTag,
		Signature:   signature2,
		PublicKey:   privateKeys[1].PublicKey(),
	}

	require.NoError(t, repo.SaveSignatureRequest(t.Context(), param2.RequestID(), signatureRequest2))
	require.NoError(t, entityProcessor.ProcessSignature(t.Context(), param2, false))

	request := entity.WantSignaturesRequest{
		WantSignatures: map[common.Hash]entity.Bitmap{
			param1.RequestID(): entity.NewBitmapOf(0), // Request validator 0 from first request
			param2.RequestID(): entity.NewBitmapOf(1), // Request validator 1 from second request
		},
	}

	response, err := syncer.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, response.Signatures, 2)
	require.Len(t, response.Signatures[param1.RequestID()], 1)
	require.Len(t, response.Signatures[param2.RequestID()], 1)
	require.Equal(t, uint32(0), response.Signatures[param1.RequestID()][0].ValidatorIndex)
	require.Equal(t, uint32(1), response.Signatures[param2.RequestID()][0].ValidatorIndex)
}

func TestHandleWantSignaturesRequest_PartialSignatureAvailability(t *testing.T) {
	t.Parallel()

	repo := createTestRepo(t)
	defer repo.Close()

	// Create test data
	validatorSet, privateKeys := createTestValidatorSetWithMultipleValidators(t, 4) // Create 4 validators
	signatureRequest := createTestSignatureRequest(t)

	nextValsetData := entity.NextValsetData{
		NextValidatorSet:     validatorSet,
		NextNetworkConfig:    randomNetworkConfig(),
		PrevValidatorSet:     validatorSet,
		PrevNetworkConfig:    randomNetworkConfig(),
		SignatureRequest:     &signatureRequest,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{},
	}

	// Setup repository
	require.NoError(t, repo.SaveNextValsetData(t.Context(), nextValsetData))

	entityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               createMockAggregator(t),
		AggProofSignal:           createMockAggProofSignal(t),
		SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
		Metrics:                  doNothingMetrics{},
	})
	require.NoError(t, err)

	var requestID common.Hash
	// Save signatures only for validator indices 0 and 2 (skip 1 and 3)
	for _, i := range []uint32{0, 2} {
		signature, hash, err := privateKeys[i].Sign(signatureRequest.Message)
		require.NoError(t, err)
		param := symbiotic.Signature{
			MessageHash: hash,
			Epoch:       signatureRequest.RequiredEpoch,
			KeyTag:      signatureRequest.KeyTag,
			Signature:   signature,
			PublicKey:   privateKeys[i].PublicKey(),
		}
		if i == 0 {
			requestID = param.RequestID()
			require.NoError(t, repo.SaveSignatureRequest(t.Context(), requestID, signatureRequest))
		}
		require.NoError(t, entityProcessor.ProcessSignature(t.Context(), param, false))
	}

	syncer, err := New(Config{
		Repo:                        repo,
		EntityProcessor:             entityProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
		MaxAggProofRequestsPerSync:  100,
		MaxResponseAggProofCount:    100,
	})
	require.NoError(t, err)

	request := entity.WantSignaturesRequest{
		WantSignatures: map[common.Hash]entity.Bitmap{
			requestID: entity.NewBitmapOf(0, 1, 2, 3), // Request all 4, but only 0 and 2 exist
		},
	}

	response, err := syncer.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, response.Signatures, 1)

	signatures := response.Signatures[requestID]
	require.Len(t, signatures, 2, "should return only available signatures")

	// Check that we got the right validator indices
	validatorIndices := make([]uint32, 0, len(signatures))
	for _, sig := range signatures {
		validatorIndices = append(validatorIndices, sig.ValidatorIndex)
	}
	require.Contains(t, validatorIndices, uint32(0))
	require.Contains(t, validatorIndices, uint32(2))
	require.NotContains(t, validatorIndices, uint32(1))
	require.NotContains(t, validatorIndices, uint32(3))
}

func randomNetworkConfig() symbiotic.NetworkConfig {
	return symbiotic.NetworkConfig{
		VerificationType:     symbiotic.VerificationTypeBlsBn254Simple,
		RequiredHeaderKeyTag: symbiotic.KeyTag(15),
		EpochDuration:        uint64(time.Minute.Seconds()),
		NumAggregators:       1,
		NumCommitters:        1,
	}
}

type doNothingMetrics struct{}

func (d doNothingMetrics) ObserveEpoch(epochType string, epochNumber uint64) {}
