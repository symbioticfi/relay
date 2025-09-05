package sync_provider

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	signature_processor "github.com/symbioticfi/relay/core/usecase/signature-processor"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
)

func TestAskSignatures_HandleWantSignaturesRequest_Integration(t *testing.T) {
	requesterRepo := createTestRepo(t)
	defer requesterRepo.Close()

	peerRepo := createTestRepo(t)
	defer peerRepo.Close()

	// Create test data
	privateKey := newPrivateKey(t)
	signatureRequest := createTestSignatureRequest(t)
	validatorSet := createTestValidatorSet(t, privateKey)

	// Setup both repositories with the same validator set
	require.NoError(t, peerRepo.SaveValidatorSet(t.Context(), validatorSet))
	require.NoError(t, requesterRepo.SaveValidatorSet(t.Context(), validatorSet))
	require.NoError(t, requesterRepo.SaveSignatureRequest(t.Context(), signatureRequest))
	signatureMap := entity.NewSignatureMap(signatureRequest.Hash(), signatureRequest.RequiredEpoch, uint32(len(validatorSet.Validators)))
	require.NoError(t, requesterRepo.UpdateSignatureMap(t.Context(), signatureMap))

	peerSignatureProcessor, err := signature_processor.NewSignatureProcessor(signature_processor.Config{
		Repo: peerRepo,
	})
	require.NoError(t, err)

	signature, hash, err := privateKey.Sign(signatureRequest.Message)
	require.NoError(t, err)

	// Save signature request and signature on peer
	param := entity.SaveSignatureParam{
		RequestHash: signatureRequest.Hash(),
		Key:         privateKey.PublicKey().Raw(), // Keep using Raw format for storage
		Signature: entity.SignatureExtended{
			MessageHash: hash,
			Signature:   signature,
			PublicKey:   privateKey.PublicKey().Raw(),
		},
		ActiveIndex:      0, // First and single validator
		VotingPower:      validatorSet.Validators[0].VotingPower,
		Epoch:            signatureRequest.RequiredEpoch,
		SignatureRequest: &signatureRequest,
	}
	require.NoError(t, peerSignatureProcessor.ProcessSignature(t.Context(), param))

	// Setup requester processor

	requesterProcessor, err := signature_processor.NewSignatureProcessor(signature_processor.Config{
		Repo: requesterRepo,
	})
	require.NoError(t, err)

	// Create peer syncer first (with a temporary mock)
	peerSyncer, err := New(Config{
		Repo:                        peerRepo,
		SignatureProcessor:          peerSignatureProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
	})
	require.NoError(t, err)

	// Create requester syncer
	requesterSyncer, err := New(Config{
		Repo:                        requesterRepo,
		SignatureProcessor:          requesterProcessor,
		EpochsToSync:                1,
		MaxSignatureRequestsPerSync: 100,
		MaxResponseSignatureCount:   100,
	})
	require.NoError(t, err)

	// Verify requester initially has no signatures
	initialSignatures, err := requesterRepo.GetAllSignatures(t.Context(), signatureRequest.Hash())
	require.NoError(t, err)
	require.Empty(t, initialSignatures)
	// Verify requester has signature request
	_, err = requesterRepo.GetSignatureRequest(t.Context(), signatureRequest.Hash())
	require.NoError(t, err)

	// Call BuildWantSignaturesRequest on requester
	request, err := requesterSyncer.BuildWantSignaturesRequest(t.Context())
	require.NoError(t, err)

	response, err := peerSyncer.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)

	stat := requesterSyncer.ProcessReceivedSignatures(t.Context(), response, request.WantSignatures)
	require.Equal(t, 0, stat.TotalErrors())

	// Verify requester now has the signature
	finalSignatures, err := requesterRepo.GetAllSignatures(t.Context(), signatureRequest.Hash())
	require.NoError(t, err)
	require.Len(t, finalSignatures, 1)

	// Verify the signature is correct
	require.Equal(t, privateKey.PublicKey().Raw(), finalSignatures[0].PublicKey)
	require.NoError(t, privateKey.PublicKey().Verify(signatureRequest.Message, finalSignatures[0].Signature))
}

func createTestRepo(t *testing.T) *badger.Repository {
	t.Helper()
	repo, err := badger.New(badger.Config{
		Dir: t.TempDir(),
	})
	require.NoError(t, err)
	return repo
}

func createTestSignatureRequest(t *testing.T) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: entity.Epoch(1),
		Message:       randomBytes(t, 100),
	}
}

func newPrivateKey(t *testing.T) crypto.PrivateKey {
	t.Helper()
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	require.NoError(t, err)

	privateKey, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, privateKeyBytes)
	require.NoError(t, err)
	return privateKey
}

func createTestValidatorSet(t *testing.T, privateKey crypto.PrivateKey) entity.ValidatorSet {
	t.Helper()
	return entity.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  entity.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: entity.ToVotingPower(big.NewInt(670)),
		Validators: []entity.Validator{{
			Operator:    common.HexToAddress("0x123"),
			VotingPower: entity.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: privateKey.PublicKey().OnChain(),
				},
			},
		}},
	}
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}
