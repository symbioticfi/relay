package aggregator_app

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/usecase/aggregator-app/mocks"
)

type testSetup struct {
	ctrl           *gomock.Controller
	mockRepo       *mocks.Mockrepository
	mockP2PClient  *mocks.Mockp2pClient
	mockAggregator *mocks.Mockaggregator
	mockMetrics    *mocks.Mockmetrics
	app            *AggregatorApp
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mocks.NewMockrepository(ctrl)
	mockP2PClient := mocks.NewMockp2pClient(ctrl)
	mockAggregator := mocks.NewMockaggregator(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	cfg := Config{
		Repo:       mockRepo,
		P2PClient:  mockP2PClient,
		Aggregator: mockAggregator,
		Metrics:    mockMetrics,
	}

	app, err := NewAggregatorApp(cfg)
	require.NoError(t, err)

	return &testSetup{
		ctrl:           ctrl,
		mockRepo:       mockRepo,
		mockP2PClient:  mockP2PClient,
		mockAggregator: mockAggregator,
		mockMetrics:    mockMetrics,
		app:            app,
	}
}

func createTestSignatureMessage() entity.SignatureMessage {
	return entity.SignatureMessage{
		RequestHash: common.HexToHash("0x123"),
		KeyTag:      entity.KeyTag(15),
		Epoch:       1,
		Signature: entity.SignatureExtended{
			MessageHash: []byte("test-message-hash"),
			PublicKey:   []byte("test-pubkey"),
			Signature:   []byte("test-signature"),
		},
	}
}

func createTestSignatureMap(thresholdReached bool, requestHash common.Hash, epoch entity.Epoch) entity.SignatureMap {
	currentVotingPower := big.NewInt(500)
	if thresholdReached {
		currentVotingPower = big.NewInt(700)
	}

	return entity.SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  epoch,
		SignedValidatorsBitmap: entity.NewSignatureBitmap(),
		CurrentVotingPower:     entity.ToVotingPower(currentVotingPower),
	}
}

func createTestValidatorSet() entity.ValidatorSet {
	return entity.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  entity.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: entity.ToVotingPower(big.NewInt(670)),
		Validators: []entity.Validator{
			{
				Operator:    common.HexToAddress("0x123"),
				VotingPower: entity.ToVotingPower(big.NewInt(1000)),
				IsActive:    true,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: entity.CompactPublicKey("test-key"),
					},
				},
			},
		},
	}
}

func TestHandleSignatureGeneratedMessage_QuorumNotReached(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup mocks for quorum not reached case
	signatureMap := createTestSignatureMap(false, msg.RequestHash, msg.Epoch) // threshold NOT reached
	validatorSet := createTestValidatorSet()

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(validatorSet, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should return nil (no error) when quorum not reached
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_QuorumReached(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup mocks for quorum reached case
	signatureMap := createTestSignatureMap(true, msg.RequestHash, msg.Epoch) // threshold reached
	validatorSet := createTestValidatorSet()
	var signatures []entity.SignatureExtended
	networkConfig := entity.NetworkConfig{
		VerificationType: entity.VerificationTypeBlsBn254Simple,
	}
	proofData := entity.AggregationProof{
		VerificationType: entity.VerificationTypeBlsBn254Simple,
	}
	stat := entity.SignatureStat{
		ReqHash: msg.RequestHash,
	}

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(validatorSet, nil)
	setup.mockRepo.EXPECT().UpdateSignatureStat(gomock.Any(), msg.RequestHash, gomock.Any(), gomock.Any()).Return(stat, nil).Times(2)
	setup.mockRepo.EXPECT().GetAllSignatures(gomock.Any(), msg.RequestHash).Return(signatures, nil)
	setup.mockRepo.EXPECT().GetConfigByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(networkConfig, nil)

	setup.mockAggregator.EXPECT().Aggregate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(proofData, nil)

	// Verify the aggregated message structure
	expectedMsg := entity.AggregatedSignatureMessage{
		RequestHash:      msg.RequestHash,
		KeyTag:           msg.KeyTag,
		Epoch:            msg.Epoch,
		AggregationProof: proofData,
	}
	setup.mockP2PClient.EXPECT().BroadcastSignatureAggregatedMessage(gomock.Any(), expectedMsg).Return(nil)

	setup.mockMetrics.EXPECT().ObserveOnlyAggregateDuration(gomock.Any())
	setup.mockMetrics.EXPECT().ObserveAppAggregateDuration(gomock.Any())
	setup.mockMetrics.EXPECT().ObserveAggCompleted(gomock.Any())

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify
	require.NoError(t, err)
}
