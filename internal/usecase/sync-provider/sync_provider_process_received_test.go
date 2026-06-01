package sync_provider

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type testRepo interface {
	repo

	SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error
}

type failingEntityProcessor struct {
	t *testing.T
}

func (p failingEntityProcessor) ProcessSignature(context.Context, symbiotic.Signature, bool) error {
	p.t.Helper()
	p.t.Fatal("unexpected signature processing")
	return nil
}

func (p failingEntityProcessor) ProcessAggregationProof(context.Context, symbiotic.AggregationProof, bool) error {
	p.t.Helper()
	p.t.Fatal("unexpected aggregation proof processing")
	return nil
}

func TestSyncer_ProcessReceivedSignatures_SkipsMismatchedRequestID(t *testing.T) {
	t.Parallel()

	for name, newRepo := range backends() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t)
			validatorSet, privateKeys := createTestValidatorSetWithMultipleValidators(t, 1)
			requestedSignatureRequest := createTestSignatureRequest(t)
			otherSignatureRequest := createTestSignatureRequest(t)

			requestedSignature := signRequest(t, privateKeys[0], requestedSignatureRequest)
			otherSignature := signRequest(t, privateKeys[0], otherSignatureRequest)
			requestedID := requestedSignature.RequestID()
			require.NotEqual(t, requestedID, otherSignature.RequestID())

			saveValidatorSetData(t, repo, validatorSet, requestedSignatureRequest, requestedID)
			syncer := newTestSyncer(t, repo, failingEntityProcessor{t: t})

			stats := syncer.ProcessReceivedSignatures(
				t.Context(),
				entity.WantSignaturesResponse{
					Signatures: map[common.Hash][]entity.ValidatorSignature{
						requestedID: {
							{
								ValidatorIndex: 0,
								Signature:      otherSignature,
							},
						},
					},
				},
				map[common.Hash]entity.Bitmap{
					requestedID: entity.NewBitmapOf(0),
				},
			)

			require.Equal(t, 0, stats.ProcessedCount)
			require.Equal(t, 1, stats.UnrequestedHashCount)

			_, err := repo.GetSignatureByIndex(t.Context(), otherSignature.RequestID(), 0)
			require.ErrorIs(t, err, entity.ErrEntityNotFound)
		})
	}
}

func TestSyncer_ProcessReceivedSignatures_SkipsMismatchedValidatorIndex(t *testing.T) {
	t.Parallel()

	for name, newRepo := range backends() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t)
			validatorSet, privateKeys := createTestValidatorSetWithMultipleValidators(t, 2)
			signatureRequest := createTestSignatureRequest(t)

			signatureFromValidator1 := signRequest(t, privateKeys[1], signatureRequest)
			requestID := signatureFromValidator1.RequestID()

			saveValidatorSetData(t, repo, validatorSet, signatureRequest, requestID)
			syncer := newTestSyncer(t, repo, failingEntityProcessor{t: t})

			stats := syncer.ProcessReceivedSignatures(
				t.Context(),
				entity.WantSignaturesResponse{
					Signatures: map[common.Hash][]entity.ValidatorSignature{
						requestID: {
							{
								ValidatorIndex: 0,
								Signature:      signatureFromValidator1,
							},
						},
					},
				},
				map[common.Hash]entity.Bitmap{
					requestID: entity.NewBitmapOf(0),
				},
			)

			require.Equal(t, 0, stats.ProcessedCount)
			require.Equal(t, 1, stats.UnrequestedSignatureCount)

			_, err := repo.GetSignatureByIndex(t.Context(), requestID, 1)
			require.ErrorIs(t, err, entity.ErrEntityNotFound)
		})
	}
}

func TestSyncer_ProcessReceivedAggregationProofs_SkipsMismatchedRequestID(t *testing.T) {
	t.Parallel()

	for name, newRepo := range backends() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t)
			syncer := newTestSyncer(t, repo, failingEntityProcessor{t: t})

			requestedID := common.BytesToHash(randomBytes(t, 32))
			proof := symbiotic.AggregationProof{
				MessageHash: randomBytes(t, 32),
				KeyTag:      symbiotic.KeyTag(15),
				Epoch:       1,
				Proof:       randomBytes(t, 96),
			}
			require.NotEqual(t, requestedID, proof.RequestID())

			stats, err := syncer.ProcessReceivedAggregationProofs(
				t.Context(),
				entity.WantAggregationProofsResponse{
					Proofs: map[common.Hash]symbiotic.AggregationProof{
						requestedID: proof,
					},
				},
				[]common.Hash{requestedID},
			)
			require.NoError(t, err)

			require.Equal(t, 0, stats.ProcessedCount)
			require.Equal(t, 1, stats.UnrequestedProofCount)

			_, err = repo.GetAggregationProof(t.Context(), proof.RequestID())
			require.ErrorIs(t, err, entity.ErrEntityNotFound)
		})
	}
}

func signRequest(t *testing.T, privateKey crypto.PrivateKey, request symbiotic.SignatureRequest) symbiotic.Signature {
	t.Helper()

	signature, messageHash, err := privateKey.Sign(request.Message)
	require.NoError(t, err)

	return symbiotic.Signature{
		MessageHash: messageHash,
		KeyTag:      request.KeyTag,
		Epoch:       request.RequiredEpoch,
		PublicKey:   privateKey.PublicKey(),
		Signature:   signature,
	}
}

func saveValidatorSetData(
	t *testing.T,
	repo testRepo,
	validatorSet symbiotic.ValidatorSet,
	signatureRequest symbiotic.SignatureRequest,
	requestID common.Hash,
) {
	t.Helper()

	require.NoError(t, repo.SaveNextValsetData(t.Context(), entity.NextValsetData{
		NextValidatorSet:     validatorSet,
		NextNetworkConfig:    randomNetworkConfig(),
		PrevValidatorSet:     validatorSet,
		PrevNetworkConfig:    randomNetworkConfig(),
		SignatureRequest:     &signatureRequest,
		ValidatorSetMetadata: symbiotic.ValidatorSetMetadata{RequestID: requestID, Epoch: validatorSet.Epoch},
	}))
}

func newTestSyncer(t *testing.T, repo repo, processor entityProcessor) *Syncer {
	t.Helper()

	syncer, err := New(Config{
		Repo:                        repo,
		EntityProcessor:             processor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 10,
		MaxResponseSignatureCount:   10,
		MaxAggProofRequestsPerSync:  10,
		MaxResponseAggProofCount:    10,
	})
	require.NoError(t, err)

	return syncer
}
