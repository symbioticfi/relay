package aggregator_app

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/entity"
	aggregationPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy"
	"github.com/symbioticfi/relay/internal/usecase/aggregator-app/mocks"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type testSetup struct {
	ctrl           *gomock.Controller
	mockRepo       *mocks.Mockrepository
	mockP2PClient  *mocks.Mockp2pClient
	mockAggregator *mocks.Mockaggregator
	mockMetrics    *mocks.Mockmetrics
	app            *AggregatorApp
	privateKey     crypto.PrivateKey
}

func newTestSetup(t *testing.T, policyType symbiotic.AggregationPolicyType, maxUnsigners uint64) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mocks.NewMockrepository(ctrl)
	mockP2PClient := mocks.NewMockp2pClient(ctrl)
	mockAggregator := mocks.NewMockaggregator(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	aggPolicy, err := aggregationPolicy.NewAggregationPolicy(policyType, maxUnsigners)
	require.NoError(t, err)

	privateKey, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	kp, err := keyprovider.NewSimpleKeystoreProvider()
	require.NoError(t, err)

	require.NoError(t, kp.AddKey(15, privateKey))

	cfg := Config{
		Repo:              mockRepo,
		P2PClient:         mockP2PClient,
		Aggregator:        mockAggregator,
		Metrics:           mockMetrics,
		AggregationPolicy: aggPolicy,
		KeyProvider:       kp,
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
		privateKey:     privateKey,
	}
}

func createTestSignatureExtended(t *testing.T, pk crypto.PrivateKey) symbiotic.SignatureExtended {
	t.Helper()
	msg := "test-message"
	sign, hash, err := pk.Sign([]byte(msg))
	require.NoError(t, err)

	return symbiotic.SignatureExtended{
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       1,
		MessageHash: hash,
		PublicKey:   pk.PublicKey().Raw(),
		Signature:   sign,
	}
}

// Unified test data structure that keeps ValidatorSet and SignatureMap in sync
type testData struct {
	ValidatorSet symbiotic.ValidatorSet
	SignatureMap entity.SignatureMap
}

// Create unified test data with a single ValidatorSet used consistently
func createTestData(requestID common.Hash, epoch symbiotic.Epoch, totalValidators, signers int, key crypto.PrivateKey) testData {
	// Create validators
	validators := make([]symbiotic.Validator, totalValidators)
	for i := 0; i < totalValidators; i++ {
		validators[i] = symbiotic.Validator{
			Operator:    common.HexToAddress(fmt.Sprintf("0x%040d", i+1)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)), // Each validator has 100 voting power
			IsActive:    true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(15),
					Payload: key.PublicKey().OnChain(),
				},
			},
		}
	}

	// Create the unified ValidatorSet
	validatorSet := symbiotic.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  symbiotic.KeyTag(15),
		Epoch:           epoch,
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(670)), // Need 670 voting power for quorum
		Validators:      validators,
		AggregatorIndices: lo.Map(validators, func(_ symbiotic.Validator, idx int) uint32 {
			return uint32(idx)
		}),
	}

	// Create SignatureMap using the same ValidatorSet
	signatureMap := entity.NewSignatureMap(requestID, epoch, uint32(totalValidators))

	// Add signers (first 'signers' number of validators)
	for i := 0; i < signers; i++ {
		votingPower := validatorSet.Validators[i].VotingPower // Use actual voting power from validator
		err := signatureMap.SetValidatorPresent(uint32(i), votingPower)
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
func createTestDataWithQuorum(requestID common.Hash, epoch symbiotic.Epoch, thresholdReached bool, key crypto.PrivateKey) testData {
	if thresholdReached {
		// 8 signers * 100 voting power = 800 > 670 threshold
		return createTestData(requestID, epoch, 10, 8, key)
	}

	// 6 signers * 100 voting power = 600 < 670 threshold
	return createTestData(requestID, epoch, 10, 6, key)
}

// Setup mocks for successful aggregation using unified test data
func setupSuccessfulAggregationMocks(setup *testSetup, msg symbiotic.SignatureExtended, testData testData) {
	var signatures []symbiotic.SignatureExtended
	networkConfig := symbiotic.NetworkConfig{
		VerificationType: symbiotic.VerificationTypeBlsBn254Simple,
	}
	proofData := symbiotic.AggregationProof{
		KeyTag:      msg.KeyTag,
		Epoch:       msg.Epoch,
		MessageHash: msg.MessageHash,
		Proof:       []byte("test-proof"),
	}

	// Use the unified test data
	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestID()).Return(testData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), msg.Epoch).Return(testData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAllSignatures(gomock.Any(), msg.RequestID()).Return(signatures, nil)
	setup.mockRepo.EXPECT().GetConfigByEpoch(gomock.Any(), msg.Epoch).Return(networkConfig, nil)
	setup.mockRepo.EXPECT().GetAggregationProof(gomock.Any(), msg.RequestID()).Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	setup.mockAggregator.EXPECT().Aggregate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(proofData, nil)

	setup.mockP2PClient.EXPECT().BroadcastSignatureAggregatedMessage(gomock.Any(), proofData).Return(nil)

	setup.mockMetrics.EXPECT().ObserveOnlyAggregateDuration(gomock.Any())
	setup.mockMetrics.EXPECT().ObserveAppAggregateDuration(gomock.Any())
}

// LOW LATENCY POLICY TESTS

func TestHandleSignatureGeneratedMessage_LowLatencyPolicy_QuorumNotReached(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowLatency, 0)
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup mocks for quorum NOT reached case
	testingData := createTestDataWithQuorum(msg.RequestID(), msg.Epoch, false, setup.privateKey)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestID()).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), msg.Epoch).Return(testingData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAggregationProof(gomock.Any(), msg.RequestID()).Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(t.Context(), msg)

	// Verify - should return nil (no error) when quorum not reached, no aggregation
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowLatencyPolicy_QuorumReached(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowLatency, 0)
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup mocks for quorum reached case - LowLatency should aggregate immediately
	testingData := createTestDataWithQuorum(msg.RequestID(), msg.Epoch, true, setup.privateKey)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should successfully aggregate when quorum reached
	require.NoError(t, err)
}

// LOW COST POLICY TESTS

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumNotReached(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 5)
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup mocks for quorum NOT reached case
	testingData := createTestDataWithQuorum(msg.RequestID(), msg.Epoch, false, setup.privateKey)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestID()).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), msg.Epoch).Return(testingData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAggregationProof(gomock.Any(), msg.RequestID()).Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should return nil (no error) when quorum not reached, no aggregation
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_TooManyUnsigners(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 2) // Allow max 2 unsigners
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 10 total validators, 7 signers = 3 unsigners (exceeds maxUnsigners=2)
	testingData := createTestData(msg.RequestID(), msg.Epoch, 10, 7, setup.privateKey)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestID()).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), msg.Epoch).Return(testingData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAggregationProof(gomock.Any(), msg.RequestID()).Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should not aggregate due to too many unsigners
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_AcceptableUnsigners(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 3) // Allow max 3 unsigners
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 10 total validators, 8 signers = 2 unsigners (within maxUnsigners=3)
	testingData := createTestData(msg.RequestID(), msg.Epoch, 10, 8, setup.privateKey)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should successfully aggregate when unsigners within limit
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_QuorumReached_ExactUnsignersLimit(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 3) // Allow max 3 unsigners
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 10 total validators, 7 signers = 3 unsigners (exactly maxUnsigners=3)
	testingData := createTestData(msg.RequestID(), msg.Epoch, 10, 7, setup.privateKey)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should successfully aggregate when exactly at unsigners limit
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_AllValidatorsSigned(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 1) // Allow max 1 unsigner
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 10 total validators, 10 signers = 0 unsigners (well within limit)
	testingData := createTestData(msg.RequestID(), msg.Epoch, 10, 10, setup.privateKey)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should successfully aggregate when all validators signed
	require.NoError(t, err)
}

// EDGE CASES

func TestHandleSignatureGeneratedMessage_LowCostPolicy_ZeroMaxUnsigners(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 0) // Allow 0 unsigners
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 5 total validators, 4 signers = 1 unsigner (exceeds maxUnsigners=0)
	testingData := createTestData(msg.RequestID(), msg.Epoch, 5, 4, setup.privateKey)

	setup.mockRepo.EXPECT().GetSignatureMap(gomock.Any(), msg.RequestID()).Return(testingData.SignatureMap, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(gomock.Any(), msg.Epoch).Return(testingData.ValidatorSet, nil)
	setup.mockRepo.EXPECT().GetAggregationProof(gomock.Any(), msg.RequestID()).Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should not aggregate due to any unsigners when maxUnsigners=0
	require.NoError(t, err)
}

func TestHandleSignatureGeneratedMessage_LowCostPolicy_HighMaxUnsigners(t *testing.T) {
	setup := newTestSetup(t, symbiotic.AggregationPolicyLowCost, 100) // Allow 100 unsigners
	ctx := context.Background()
	msg := createTestSignatureExtended(t, setup.privateKey)

	// Setup: 10 total validators, 7 signers = 3 unsigners (well within limit)
	// 7 signers = 7*100 = 700 > 670 for quorum
	testingData := createTestData(msg.RequestID(), msg.Epoch, 10, 7, setup.privateKey)

	setupSuccessfulAggregationMocks(setup, msg, testingData)

	// Execute
	err := setup.app.HandleSignatureProcessedMessage(ctx, msg)

	// Verify - should successfully aggregate with high unsigners limit
	require.NoError(t, err)
}

// Test helper function to verify SignatureMap functionality with unified test data
func TestSignatureMapFunctionality(t *testing.T) {
	requestID := common.HexToHash("0x123")

	// Test with unified creation
	pk, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	testingData := createTestData(requestID, 1, 5, 0, pk) // 5 validators, 0 signers initially
	signatureMap := testingData.SignatureMap
	validatorSet := testingData.ValidatorSet

	// Initially no validators signed
	require.Equal(t, uint64(0), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold))

	// Add 3 validators using their actual voting power from the validator set
	for i := 0; i < 3; i++ {
		votingPower := validatorSet.Validators[i].VotingPower
		err := signatureMap.SetValidatorPresent(uint32(i), votingPower)
		require.NoError(t, err)
	}

	require.Equal(t, uint64(3), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold)) // 300 < 670

	// Add 4 more validators (5 total = 5 * 100 = 500 voting power)
	for i := 3; i < 5; i++ {
		votingPower := validatorSet.Validators[i].VotingPower
		err := signatureMap.SetValidatorPresent(uint32(i), votingPower)
		require.NoError(t, err)
	}

	require.Equal(t, uint64(5), signatureMap.SignedValidatorsBitmap.GetCardinality())
	require.False(t, signatureMap.ThresholdReached(validatorSet.QuorumThreshold))            // 500 < 670
	require.True(t, signatureMap.ThresholdReached(symbiotic.ToVotingPower(big.NewInt(400)))) // 500 >= 400

	// Verify that the SignatureMap and ValidatorSet are consistent
	require.Equal(t, int64(5), validatorSet.GetTotalActiveValidators())
	require.Equal(t, validatorSet.QuorumThreshold, symbiotic.ToVotingPower(big.NewInt(670)))
}
