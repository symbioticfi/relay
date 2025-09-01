package aggregator_app

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	aggregationPolicy2 "github.com/symbioticfi/relay/internal/usecase/aggregation-policy"

	"github.com/RoaringBitmap/roaring/v2"
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

func newTestSetup(t *testing.T, policyType entity.AggregationPolicyType, maxUnsigners uint64) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mocks.NewMockrepository(ctrl)
	mockP2PClient := mocks.NewMockp2pClient(ctrl)
	mockAggregator := mocks.NewMockaggregator(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	aggregationPolicy, err := aggregationPolicy2.NewAggregationPolicy(policyType, maxUnsigners)
	require.NoError(t, err)

	cfg := Config{
		Repo:              mockRepo,
		P2PClient:         mockP2PClient,
		Aggregator:        mockAggregator,
		Metrics:           mockMetrics,
		AggregationPolicy: aggregationPolicy,
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

// Enhanced signature map creation with proper threshold and signing control
func createTestSignatureMapWithSigners(thresholdReached bool, requestHash common.Hash, epoch uint64, totalValidators, signers int) entity.SignatureMap {
	activeValidatorsMap := make(map[common.Address]struct{})
	isPresent := make(map[common.Address]struct{})

	// Create validators
	for i := 0; i < totalValidators; i++ {
		validatorAddr := common.HexToAddress(fmt.Sprintf("0x%040d", i+1))

		activeValidatorsMap[validatorAddr] = struct{}{}

		// Mark first 'signers' number of validators as having signed
		if i < signers {
			isPresent[validatorAddr] = struct{}{}
		}
	}

	currentVotingPower := big.NewInt(int64(signers * 100)) // signers * 100 voting power each
	totalVotingPower := big.NewInt(int64(totalValidators * 100))
	quorumThreshold := big.NewInt(670) // Default threshold

	if thresholdReached {
		// Ensure current voting power exceeds threshold
		if currentVotingPower.Cmp(quorumThreshold) <= 0 {
			currentVotingPower = big.NewInt(700)
		}
	} else {
		// Ensure current voting power is below threshold
		if currentVotingPower.Cmp(quorumThreshold) >= 0 {
			currentVotingPower = big.NewInt(600)
		}
	}

	return entity.SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  epoch,
		SignedValidatorsBitmap: roaring.New(),
		CurrentVotingPower:     entity.ToVotingPower(currentVotingPower),
	}
}

func createTestSignatureMap(thresholdReached bool, requestHash common.Hash, epoch uint64) entity.SignatureMap {
	return createTestSignatureMapWithSigners(thresholdReached, requestHash, epoch, 10, 7)
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

// Setup mocks for successful aggregation
func setupSuccessfulAggregationMocks(setup *testSetup, msg entity.SignatureMessage, signatureMap entity.SignatureMap) {
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
	setup.mockRepo.EXPECT().UpdateSignatureStat(gomock.Any(), msg.RequestHash, gomock.Any(), gomock.Any()).Return(stat, nil).Times(2)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(validatorSet, nil)
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
}

// LOW LATENCY POLICY TESTS

func TestHandleSignatureGeneratedMessage_LowLatencyPolicy_QuorumNotReached(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowLatency, 0)
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup mocks for quorum NOT reached case
	signatureMap := createTestSignatureMap(false, msg.RequestHash, uint64(msg.Epoch))

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should return nil (no error) when quorum not reached, no aggregation
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowLatencyPolicy_QuorumReached(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowLatency, 0)
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup mocks for quorum reached case - LowLatency should aggregate immediately
	validatorMap := createTestSignatureMap(true, msg.RequestHash, uint64(msg.Epoch))

	setupSuccessfulAggregationMocks(setup, msg, validatorMap)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate when quorum reached
	require.NoError(t, err)
}

// LOW COST POLICY TESTS

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumNotReached(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 5)
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup mocks for quorum NOT reached case
	signatureMap := createTestSignatureMap(false, msg.RequestHash, uint64(msg.Epoch))

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should return nil (no error) when quorum not reached, no aggregation
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_TooManyUnsigners(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 2) // Allow max 2 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 7 signers = 3 unsigners (exceeds maxUnsigners=2)
	signatureMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 10, 7)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should not aggregate due to too many unsigners
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_AcceptableUnsigners(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 3) // Allow max 3 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 8 signers = 2 unsigners (within maxUnsigners=3)
	validatorMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 10, 8)

	setupSuccessfulAggregationMocks(setup, msg, validatorMap)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate when unsigners within limit
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_ExactUnsignersLimit(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 3) // Allow max 3 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 7 signers = 3 unsigners (exactly maxUnsigners=3)
	validatorMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 10, 7)

	setupSuccessfulAggregationMocks(setup, msg, validatorMap)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate when exactly at unsigners limit
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_AllValidatorsSigned(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 1) // Allow max 1 unsigner
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 10 signers = 0 unsigners (well within limit)
	validatorMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 10, 10)

	setupSuccessfulAggregationMocks(setup, msg, validatorMap)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate when all validators signed
	require.NoError(t, err)
}

// EDGE CASES

func TestHandleSignatureGeneratedMessage_LowCostPolicy_ZeroMaxUnsigners(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 0) // Allow 0 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 5 total validators, 4 signers = 1 unsigner (exceeds maxUnsigners=0)
	signatureMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 5, 4)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(signatureMap, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should not aggregate due to any unsigners when maxUnsigners=0
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_HighMaxUnsigners(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 100) // Allow 100 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 5 signers = 5 unsigners (well within limit)
	validatorMap := createTestSignatureMapWithSigners(true, msg.RequestHash, uint64(msg.Epoch), 10, 5)

	setupSuccessfulAggregationMocks(setup, msg, validatorMap)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate with high unsigners limit
	require.NoError(t, err)
}
