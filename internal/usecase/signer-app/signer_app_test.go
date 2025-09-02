package signer_app

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	signature_processor "github.com/symbioticfi/relay/core/usecase/signature-processor"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
	"github.com/symbioticfi/relay/internal/usecase/signer-app/mocks"
)

func TestSign_HappyPath(t *testing.T) {
	setup := newTestSetup(t)
	msg := "test-message-to-sign"
	req := createTestSignatureRequest(msg)
	privateKey := newPrivateKey(t)
	createTestValidatorSet(t, setup, privateKey)

	// Add the private key to the real key provider
	require.NoError(t, setup.keyProvider.AddKey(req.KeyTag, privateKey))

	// Mock the remaining dependencies
	setup.mockP2P.EXPECT().BroadcastSignatureGeneratedMessage(gomock.Any(), gomock.Any()).Return(nil)
	setup.mockMetrics.EXPECT().ObservePKSignDuration(gomock.Any())
	setup.mockMetrics.EXPECT().ObserveAppSignDuration(gomock.Any())

	// Sign
	require.NoError(t, setup.app.Sign(t.Context(), req))

	// Verify that signature request was saved
	savedReq, err := setup.repo.GetSignatureRequest(t.Context(), req.Hash())
	require.NoError(t, err)
	require.Equal(t, req.KeyTag, savedReq.KeyTag)
	require.Equal(t, req.RequiredEpoch, savedReq.RequiredEpoch)
	require.Equal(t, req.Message, savedReq.Message)

	// Verify that signature is correct
	signatures, err := setup.repo.GetAllSignatures(t.Context(), req.Hash())
	require.NoError(t, err)
	require.Len(t, signatures, 1)

	require.NoError(t, privateKey.PublicKey().Verify([]byte(msg), signatures[0].Signature))
}

type testSetup struct {
	ctrl           *gomock.Controller
	repo           *badger.Repository
	keyProvider    *keyprovider.SimpleKeystoreProvider
	mockP2P        *mocks.Mockp2pService
	mockAggProof   *mocks.MockaggProofSignal
	mockAggregator *mocks.Mockaggregator
	mockMetrics    *mocks.Mockmetrics
	app            *SignerApp
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)

	repo, err := badger.New(badger.Config{
		Dir: t.TempDir(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.Close()
	})

	keyProvider, err := keyprovider.NewSimpleKeystoreProvider()
	require.NoError(t, err)

	// Create mocks for other dependencies
	mockP2P := mocks.NewMockp2pService(ctrl)
	mockAggProof := mocks.NewMockaggProofSignal(ctrl)
	mockAggregator := mocks.NewMockaggregator(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	processor, err := signature_processor.NewSignatureProcessor(signature_processor.Config{
		Repo: repo,
	})
	require.NoError(t, err)

	cfg := Config{
		P2PService:         mockP2P,
		KeyProvider:        keyProvider,
		Repo:               repo,
		SignatureProcessor: processor,
		AggProofSignal:     mockAggProof,
		Aggregator:         mockAggregator,
		Metrics:            mockMetrics,
	}

	app, err := NewSignerApp(cfg)
	require.NoError(t, err)

	return &testSetup{
		ctrl:           ctrl,
		repo:           repo,
		mockP2P:        mockP2P,
		keyProvider:    keyProvider,
		mockAggProof:   mockAggProof,
		mockAggregator: mockAggregator,
		mockMetrics:    mockMetrics,
		app:            app,
	}
}

func createTestSignatureRequest(msg string) entity.SignatureRequest {
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: entity.Epoch(1),
		Message:       []byte(msg),
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

func createTestValidatorSet(t *testing.T, setup *testSetup, privateKey crypto.PrivateKey) entity.ValidatorSet {
	t.Helper()
	vs := entity.ValidatorSet{
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

	require.NoError(t, setup.repo.SaveValidatorSet(t.Context(), vs))

	return vs
}
