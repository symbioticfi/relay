package signer_app

import (
	"crypto/rand"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/client/repository/badger"
	entity_processor "github.com/symbioticfi/relay/internal/usecase/entity-processor"
	entity_mocks "github.com/symbioticfi/relay/internal/usecase/entity-processor/mocks"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/signer-app/mocks"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestSign_HappyPath(t *testing.T) {
	setup := newTestSetup(t)
	req := createTestSignatureRequest(lo.RandomString(100, lo.AllCharset))
	privateKey := newPrivateKey(t)
	createTestValidatorSet(t, setup, privateKey)

	// Add the private key to the real key provider
	require.NoError(t, setup.keyProvider.AddKey(req.KeyTag, privateKey))

	// Mock the remaining dependencies
	setup.mockP2P.EXPECT().BroadcastSignatureGeneratedMessage(gomock.Any(), gomock.Any()).Return(nil)
	setup.mockMetrics.EXPECT().ObservePKSignDuration(gomock.Any())
	setup.mockMetrics.EXPECT().ObserveAppSignDuration(gomock.Any())

	go setup.app.HandleSignatureRequests(t.Context(), 1, setup.mockP2P)

	// Sign
	reqID, err := setup.app.RequestSignature(t.Context(), req)
	require.NoError(t, err)

	// Verify that signature request was saved
	savedReq, err := setup.repo.GetSignatureRequest(t.Context(), reqID)
	require.NoError(t, err)
	require.Equal(t, req.KeyTag, savedReq.KeyTag)
	require.Equal(t, req.RequiredEpoch, savedReq.RequiredEpoch)
	require.Equal(t, req.Message, savedReq.Message)

	time.Sleep(time.Second)

	// Verify that signature is correct
	signatures, err := setup.repo.GetAllSignatures(t.Context(), reqID)
	require.NoError(t, err)
	require.Len(t, signatures, 1)

	require.NoError(t, privateKey.PublicKey().Verify(req.Message, signatures[0].Signature))
}

type testSetup struct {
	ctrl        *gomock.Controller
	repo        *badger.Repository
	keyProvider *keyprovider.SimpleKeystoreProvider
	mockP2P     *mocks.Mockp2pService
	mockMetrics *mocks.Mockmetrics
	app         *SignerApp
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctrl := gomock.NewController(t)

	repo, err := badger.New(badger.Config{
		Dir:     t.TempDir(),
		Metrics: badger.DoNothingMetrics{},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.Close()
	})

	keyProvider, err := keyprovider.NewSimpleKeystoreProvider()
	require.NoError(t, err)

	// Create mocks for other dependencies
	mockP2P := mocks.NewMockp2pService(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	// Create mock aggregator for entity processor
	mockEntityAggregator := entity_mocks.NewMockAggregator(ctrl)
	mockEntityAggregator.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	// Create mock aggregation proof signal for entity processor
	mockEntityAggProofSignal := entity_mocks.NewMockAggProofSignal(ctrl)
	mockEntityAggProofSignal.EXPECT().Emit(gomock.Any()).Return(nil).AnyTimes()

	// Create mock signature processed signal for entity processor
	signatureProcessedSignal := signals.New[symbiotic.Signature](signals.DefaultConfig(), "test", nil)

	processor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               mockEntityAggregator,
		AggProofSignal:           mockEntityAggProofSignal,
		SignatureProcessedSignal: signatureProcessedSignal,
	})
	require.NoError(t, err)

	cfg := Config{
		KeyProvider:     keyprovider.NewCacheKeyProvider(keyProvider),
		Repo:            repo,
		EntityProcessor: processor,
		Metrics:         mockMetrics,
	}

	app, err := NewSignerApp(cfg)
	require.NoError(t, err)

	return &testSetup{
		ctrl:        ctrl,
		repo:        repo,
		mockP2P:     mockP2P,
		keyProvider: keyProvider,
		mockMetrics: mockMetrics,
		app:         app,
	}
}

func createTestSignatureRequest(msg string) symbiotic.SignatureRequest {
	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: symbiotic.Epoch(1),
		Message:       []byte(msg),
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

func createTestValidatorSet(t *testing.T, setup *testSetup, privateKey crypto.PrivateKey) symbiotic.ValidatorSet {
	t.Helper()
	vs := symbiotic.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  symbiotic.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(670)),
		Validators: []symbiotic.Validator{{
			Operator:    common.HexToAddress("0x123"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(15),
					Payload: privateKey.PublicKey().OnChain(),
				},
			},
		}},
	}

	require.NoError(t, setup.repo.SaveValidatorSet(t.Context(), vs))

	return vs
}
