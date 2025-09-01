package aggregator_app

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	aggregationPolicy2 "github.com/symbioticfi/relay/internal/usecase/aggregation-policy"

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

// Unified test data structure that keeps ValidatorSet and SignatureMap in sync
type testData struct {
	ValidatorSet entity.ValidatorSet
	SignatureMap entity.SignatureMap
}

// Create unified test data with a single ValidatorSet used consistently
func createTestData(requestHash common.Hash, epoch uint64, totalValidators, signers int) testData {
	// Create validators
	validators := make([]entity.Validator, totalValidators)
	for i := 0; i < totalValidators; i++ {
		validators[i] = entity.Validator{
			Operator:    common.HexToAddress(fmt.Sprintf("0x%040d", i+1)),
			VotingPower: entity.ToVotingPower(big.NewInt(100)), // Each validator has 100 voting power
			IsActive:    true,
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: entity.CompactPublicKey(fmt.Sprintf("test-key-%d", i+1)),
				},
			},
		}
	}

	// Create the unified ValidatorSet
	validatorSet := entity.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  entity.KeyTag(15),
		Epoch:           epoch,
		QuorumThreshold: entity.ToVotingPower(big.NewInt(670)), // Need 670 voting power for quorum
		Validators:      validators,
	}

	// Create SignatureMap using the same ValidatorSet
	signatureMap := entity.NewSignatureMap(requestHash, validatorSet)

	// Add signers (first 'signers' number of validators)
	for i := 0; i < signers; i++ {
		votingPower := validatorSet.Validators[i].VotingPower // Use actual voting power from validator
		err := signatureMap.SetValidatorPresent(i, votingPower)
		if err != nil {
			panic(fmt.Sprintf("Failed to set validator present: %v", err))
		}
	}

	return testData{
		ValidatorSet: validatorSet,
		SignatureMap: signatureMap,
	}
}

// Convenience function for common test scenarios
func createTestDataWithQuorum(requestHash common.Hash, epoch uint64, thresholdReached bool) testData {
	if thresholdReached {
		// 8 signers * 100 voting power = 800 > 670 threshold
		return createTestData(requestHash, epoch, 10, 8)
	}

	// 6 signers * 100 voting power = 600 < 670 threshold
	return createTestData(requestHash, epoch, 10, 6)
}

// Setup mocks for successful aggregation using unified test data
func setupSuccessfulAggregationMocks(setup *testSetup, msg entity.SignatureMessage, testData testData) {
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

	// Use the unified test data
	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(testData.SignatureMap, nil)
	setup.mockRepo.EXPECT().UpdateSignatureStat(gomock.Any(), msg.RequestHash, gomock.Any(), gomock.Any()).Return(stat, nil).Times(2)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(testData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAllSignatures(gomock.Any(), msg.RequestHash).Return(signatures, nil)
	setup.mockRepo.EXPECT().GetConfigByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(networkConfig, nil)

	setup.mockAggregator.EXPECT().Aggregate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(proofData, nil)

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
	testingData := createTestDataWithQuorum(msg.RequestHash, uint64(msg.Epoch), false)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(testingData.ValidatorSet, nil)

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
	testingData := createTestDataWithQuorum(msg.RequestHash, uint64(msg.Epoch), true)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

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
	testingData := createTestDataWithQuorum(msg.RequestHash, uint64(msg.Epoch), false)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(testingData.ValidatorSet, nil)

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
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 10, 7)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(testingData.ValidatorSet, nil)

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
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 10, 8)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

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
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 10, 7)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

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
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 10, 10)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

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
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 5, 4)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestHash).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), uint64(msg.Epoch)).Return(testingData.ValidatorSet, nil)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should not aggregate due to any unsigners when maxUnsigners=0
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_HighMaxUnsigners(t *testing.T) {
	setup := newTestSetup(t, entity.AggregationPolicyLowCost, 100) // Allow 100 unsigners
	ctx := context.Background()
	msg := createTestSignatureMessage()

	// Setup: 10 total validators, 7 signers = 3 unsigners (well within limit)
	// 7 signers = 7*100 = 700 > 670 for quorum
	testingData := createTestData(msg.RequestHash, uint64(msg.Epoch), 10, 7)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureGeneratedMessage(ctx, msg)

	// Verify - should successfully aggregate with high unsigners limit
	require.NoError(t, err)
}

// Test helper function to verify SignatureMap functionality with unified test data
func TestSignatureMapFunctionality(t *testing.T) {
	requestHash := common.HexToHash("0x123")

	// Test with unified creation
	testingData := createTestData(requestHash, 1, 5, 0) // 5 validators, 0 signers initially
	signatureMap := testingData.SignatureMap
	validatorSet := testingData.ValidatorSet

	// Initially no validators signed
	require.Equal(t, uint64(0), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold))

	// Add 3 validators using their actual voting power from the validator set
	for i := 0; i < 3; i++ {
		votingPower := validatorSet.Validators[i].VotingPower
		err := signatureMap.SetValidatorPresent(i, votingPower)
		require.NoError(t, err)
	}

	require.Equal(t, uint64(3), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold)) // 300 < 670

	// Add 4 more validators (5 total = 5 * 100 = 500 voting power)
	for i := 3; i < 5; i++ {
		votingPower := validatorSet.Validators[i].VotingPower
		err := signatureMap.SetValidatorPresent(i, votingPower)
		require.NoError(t, err)
	}

	require.Equal(t, uint64(5), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold))         // 500 < 670
	require.True(t, signatureMap.ThresholdReached(entity.ToVotingPower(big.NewInt(400)))) // 500 >= 400

	// Verify that the SignatureMap and ValidatorSet are consistent
	require.Equal(t, int64(5), validatorSet.GetTotalActiveValidators())
	require.Equal(t, validatorSet.QuorumThreshold, entity.ToVotingPower(big.NewInt(670)))
}
