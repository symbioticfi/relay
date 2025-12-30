package api_server

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetCustomScheduleNodeStatus(t *testing.T) {
	t.Run("Success_WithSpecificEpoch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		requestedEpoch := symbiotic.Epoch(5)
		currentEpoch := symbiotic.Epoch(10)
		localKey := symbiotic.CompactPublicKey("local-validator-key")

		validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
		validatorSet.Validators[0].Keys[0].Payload = localKey

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

		slotDuration := uint64(60) // 60 seconds per slot
		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  (*uint64)(&requestedEpoch),
			Seed:                   []byte("test-seed"),
			SlotDurationSeconds:    slotDuration,
			MaxParticipantsPerSlot: 1,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		// IsActive can be true or false depending on the schedule
	})

	t.Run("Success_UseCurrentEpoch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		currentEpoch := symbiotic.Epoch(10)
		localKey := symbiotic.CompactPublicKey("local-validator-key")

		validatorSet := createTestValidatorSetWithMultipleValidators(currentEpoch)
		validatorSet.Validators[0].Keys[0].Payload = localKey

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
	})

	t.Run("InvalidInput_ZeroSlotDuration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &grpcHandler{
			cfg: Config{},
		}

		ctx := context.Background()

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			SlotDurationSeconds:    0, // Invalid
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "slot_duration_seconds must be greater than 0")
	})

	t.Run("InvalidInput_ZeroMaxParticipants", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &grpcHandler{
			cfg: Config{},
		}

		ctx := context.Background()

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 0, // Invalid
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "max_participants_per_slot must be greater than 0")
	})

	t.Run("InvalidInput_MinGreaterThanMax", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &grpcHandler{
			cfg: Config{},
		}

		ctx := context.Background()

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 5, // Greater than max
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "min_participants_per_slot cannot be greater than max_participants_per_slot")
	})

	t.Run("InvalidInput_EpochFromFuture", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo: mockRepo,
			},
		}

		ctx := context.Background()
		currentEpoch := symbiotic.Epoch(10)
		futureEpoch := uint64(15)

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  &futureEpoch,
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "is greater than latest epoch")
	})

	t.Run("Error_NoActiveValidators", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		requestedEpoch := symbiotic.Epoch(5)
		currentEpoch := symbiotic.Epoch(10)
		localKey := symbiotic.CompactPublicKey("local-validator-key")

		validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
		// Mark all validators as inactive
		for i := range validatorSet.Validators {
			validatorSet.Validators[i].IsActive = false
		}

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  (*uint64)(&requestedEpoch),
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "no active validators found")
	})

	t.Run("Success_ActiveNodeHasSlotTimes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		requestedEpoch := symbiotic.Epoch(5)
		currentEpoch := symbiotic.Epoch(10)
		localKey := symbiotic.CompactPublicKey("local-validator-key")

		validatorSet := createTestValidatorSet(requestedEpoch)
		validatorSet.Validators[0].Keys[0].Payload = localKey

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  (*uint64)(&requestedEpoch),
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 1,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.True(t, response.GetIsActive())

		// Verify slot times are populated
		require.NotNil(t, response.GetCurrentSlotStartTime())
		require.NotNil(t, response.GetCurrentSlotEndTime())

		// Verify slot duration is correct
		slotStart := response.GetCurrentSlotStartTime().AsTime()
		slotEnd := response.GetCurrentSlotEndTime().AsTime()
		require.Equal(t, 60*time.Second, slotEnd.Sub(slotStart))
	})

	t.Run("Success_LocalValidatorNotActive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		requestedEpoch := symbiotic.Epoch(5)
		currentEpoch := symbiotic.Epoch(10)
		localKey := symbiotic.CompactPublicKey("local-validator-key")

		validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
		// Set local key to an inactive validator
		validatorSet.Validators[2].Keys[0].Payload = localKey

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  (*uint64)(&requestedEpoch),
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.False(t, response.GetIsActive()) // Should return false, not error

		// Verify slot times are still populated even when inactive (showing current slot)
		require.NotNil(t, response.GetCurrentSlotStartTime())
		require.NotNil(t, response.GetCurrentSlotEndTime())
	})

	t.Run("Error_KeyProviderFailure", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockrepo(ctrl)
		mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

		handler := &grpcHandler{
			cfg: Config{
				Repo:        mockRepo,
				KeyProvider: mockKeyProvider,
			},
		}

		ctx := context.Background()
		requestedEpoch := symbiotic.Epoch(5)
		currentEpoch := symbiotic.Epoch(10)
		expectedError := errors.New("key provider error")

		validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)

		mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
		mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
		mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(symbiotic.CompactPublicKey(""), expectedError)

		req := &apiv1.GetCustomScheduleNodeStatusRequest{
			Epoch:                  (*uint64)(&requestedEpoch),
			SlotDurationSeconds:    60,
			MaxParticipantsPerSlot: 2,
			MinParticipantsPerSlot: 1,
		}

		response, err := handler.GetCustomScheduleNodeStatus(ctx, req)

		require.Error(t, err)
		require.Nil(t, response)
		require.Contains(t, err.Error(), "failed to get onchain key from cache")
	})
}

// Test the schedule algorithm logic
func TestIsNodeActiveInCustomSchedule_SingleValidator(t *testing.T) {
	validatorSet := createTestValidatorSet(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	localAddress := activeValidators[0].Operator

	currentSlot := uint64(0) // First slot

	isActive, err := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		currentSlot,
		1, // max 1 per slot
		1, // min 1 per slot
		localAddress,
	)

	require.NoError(t, err)
	// With only 1 active validator, it should always be active in slot 0
	require.True(t, isActive)
}

func TestIsNodeActiveInCustomSchedule_MultipleValidators(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	firstValidator := activeValidators[0].Operator

	// Test that the validator is active in exactly one slot per cycle
	foundActiveSlot := -1
	for slotNum := 0; slotNum < 3; slotNum++ {
		isActive, err := isNodeActiveInCustomSchedule(
			activeValidators,
			symbiotic.Epoch(5),
			[]byte("test-seed"),
			uint64(slotNum),
			1, // max 1 per slot
			1, // min 1 per slot
			firstValidator,
		)

		require.NoError(t, err)
		if isActive {
			foundActiveSlot = slotNum
			break
		}
	}

	// Should find at least one active slot
	require.GreaterOrEqual(t, foundActiveSlot, 0, "Validator should be active in at least one slot")
}

func TestGetCurrentSlot_BeforeEpochStart(t *testing.T) {
	validatorSet := createTestValidatorSet(symbiotic.Epoch(5))
	epochStart := time.Unix(int64(validatorSet.CaptureTimestamp), 0)
	currentTimeFunc := func() time.Time {
		return epochStart.Add(-30 * time.Second) // Before epoch start
	}

	currentSlot, slotStart, err := getCurrentSlot(
		epochStart,
		currentTimeFunc,
		60,
	)

	// Should return an error when current time is before epoch start
	require.Error(t, err)
	require.Contains(t, err.Error(), "Epoch has not started yet")
	require.Equal(t, uint64(0), currentSlot)
	require.True(t, slotStart.IsZero())
}

func TestIsNodeActiveInCustomSchedule_NotInValidatorSet(t *testing.T) {
	validatorSet := createTestValidatorSet(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	nonExistentAddress := common.HexToAddress("0x9999999999999999999999999999999999999999")

	currentSlot := uint64(0) // First slot

	isActive, err := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		currentSlot,
		1,
		1,
		nonExistentAddress,
	)

	require.NoError(t, err)
	// Should not be active if address is not in validator set
	require.False(t, isActive)
}

func TestIsNodeActiveInCustomSchedule_DifferentSeeds(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	localAddress := activeValidators[0].Operator

	currentSlot := uint64(0) // First slot

	seed1 := []byte("seed-1")
	seed2 := []byte("seed-2")

	isActive1, err1 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		seed1,
		currentSlot,
		1,
		1,
		localAddress,
	)

	isActive2, err2 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		seed2,
		currentSlot,
		1,
		1,
		localAddress,
	)

	require.NoError(t, err1)
	require.NoError(t, err2)
	// Different seeds should potentially produce different schedules
	require.False(t, isActive1)
	require.True(t, isActive2)
}

func TestIsNodeActiveInCustomSchedule_GroupCycling(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	require.Len(t, activeValidators, 2, "Should have 2 active validators")

	firstValidator := activeValidators[0].Operator

	// With 2 validators and max 1 per slot, we have 2 groups
	// Groups should cycle: slot 0 -> group 0, slot 1 -> group 1, slot 2 -> group 0, etc.

	active0, err0 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		0, // Slot 0
		1,
		1,
		firstValidator,
	)

	active2, err2 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		2, // Slot 2 (should cycle back)
		1,
		1,
		firstValidator,
	)

	require.NoError(t, err0)
	require.NoError(t, err2)
	// Slot 0 and slot 2 should have the same status (cycling)
	require.Equal(t, active0, active2, "Groups should cycle with modulo")
}

func TestIsNodeActiveInCustomSchedule_RemainderGroup(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	require.Len(t, activeValidators, 2, "Should have 2 active validators")

	currentSlot := uint64(0) // First slot

	// With 2 validators, max 3, min 1: should create 1 group with 2 validators
	// (remainder 2 >= min 1, but since it's less than a full group, all go in one group)
	foundActive := false
	for _, val := range activeValidators {
		isActive, err := isNodeActiveInCustomSchedule(
			activeValidators,
			symbiotic.Epoch(5),
			[]byte("test-seed"),
			currentSlot,
			3, // max 3 per slot
			1, // min 1 per slot
			val.Operator,
		)
		require.NoError(t, err)
		if isActive {
			foundActive = true
		}
	}

	// At least one validator should be active in slot 0
	require.True(t, foundActive, "At least one validator should be active")
}

func TestIsNodeActiveInCustomSchedule_MultipleValidatorsInSameGroup(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	require.Len(t, activeValidators, 2, "Should have 2 active validators")

	currentSlot := uint64(0) // First slot

	// With 2 validators, max 2, min 1: should create 1 group with both validators
	validator1Active, err1 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		currentSlot,
		2, // max 2 per slot - both validators fit in one group
		1, // min 1 per slot
		activeValidators[0].Operator,
	)
	require.NoError(t, err1)

	validator2Active, err2 := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		currentSlot,
		2, // max 2 per slot - both validators fit in one group
		1, // min 1 per slot
		activeValidators[1].Operator,
	)
	require.NoError(t, err2)

	// BOTH validators should be active in the same slot since they're in the same group
	require.True(t, validator1Active, "First validator should be active in slot 0")
	require.True(t, validator2Active, "Second validator should be active in slot 0")
}

func TestIsNodeActiveInCustomSchedule_TooFewValidators(t *testing.T) {
	validatorSet := createTestValidatorSetWithMultipleValidators(symbiotic.Epoch(5))
	activeValidators := validatorSet.Validators.GetActiveValidators()
	require.Len(t, activeValidators, 2, "Should have 2 active validators")

	currentSlot := uint64(0) // First slot

	// With 2 validators, max 5, min 3: no valid groups can be formed
	// (remainder 2 < min 3)
	isActive, err := isNodeActiveInCustomSchedule(
		activeValidators,
		symbiotic.Epoch(5),
		[]byte("test-seed"),
		currentSlot,
		5, // max 5 per slot
		3, // min 3 per slot
		activeValidators[0].Operator,
	)

	// Should return an error when no valid groups can be formed
	require.Error(t, err)
	require.Contains(t, err.Error(), "no valid groups can be formed")
	require.False(t, isActive, "Should not be active when group constraints cannot be met")
}

func TestGetCurrentSlot(t *testing.T) {
	t.Run("Success_SlotCalculation", func(t *testing.T) {
		validatorSet := createTestValidatorSet(symbiotic.Epoch(5))
		epochStart := time.Unix(int64(validatorSet.CaptureTimestamp), 0)

		// Test slot 0 (0-60 seconds)
		currentTimeFunc := func() time.Time {
			return epochStart.Add(30 * time.Second)
		}

		currentSlot, slotStart, err := getCurrentSlot(
			epochStart,
			currentTimeFunc,
			60,
		)

		require.NoError(t, err)
		require.Equal(t, uint64(0), currentSlot)
		require.Equal(t, epochStart, slotStart)
	})

	t.Run("Success_SlotCalculation_SecondSlot", func(t *testing.T) {
		validatorSet := createTestValidatorSet(symbiotic.Epoch(5))
		epochStart := time.Unix(int64(validatorSet.CaptureTimestamp), 0)

		// Test slot 1 (60-120 seconds)
		currentTimeFunc := func() time.Time {
			return epochStart.Add(90 * time.Second)
		}

		currentSlot, slotStart, err := getCurrentSlot(
			epochStart,
			currentTimeFunc,
			60,
		)

		require.NoError(t, err)
		require.Equal(t, uint64(1), currentSlot)
		require.Equal(t, epochStart.Add(60*time.Second), slotStart)
	})
}

func TestCreateSeededRNG(t *testing.T) {
	t.Run("Deterministic_SameSeedProducesSameSequence", func(t *testing.T) {
		epoch := symbiotic.Epoch(5)
		seed := []byte("test-seed")

		rng1 := createSeededRNG(epoch, seed)
		rng2 := createSeededRNG(epoch, seed)

		// Same seed should produce same sequence
		for i := 0; i < 10; i++ {
			require.Equal(t, rng1.Intn(100), rng2.Intn(100))
		}
	})

	t.Run("DifferentSeeds_ProduceDifferentSequences", func(t *testing.T) {
		epoch := symbiotic.Epoch(5)
		seed1 := []byte("seed-1")
		seed2 := []byte("seed-2")

		rng1 := createSeededRNG(epoch, seed1)
		rng2 := createSeededRNG(epoch, seed2)

		// Different seeds should produce different sequences (with high probability)
		vals1 := make([]int, 10)
		vals2 := make([]int, 10)
		for i := 0; i < 10; i++ {
			vals1[i] = rng1.Intn(1000)
			vals2[i] = rng2.Intn(1000)
		}

		require.NotEqual(t, vals1, vals2)
	})

	t.Run("DifferentEpochs_ProduceDifferentSequences", func(t *testing.T) {
		epoch1 := symbiotic.Epoch(5)
		epoch2 := symbiotic.Epoch(10)
		seed := []byte("test-seed")

		rng1 := createSeededRNG(epoch1, seed)
		rng2 := createSeededRNG(epoch2, seed)

		// Different epochs should produce different sequences (with high probability)
		vals1 := make([]int, 10)
		vals2 := make([]int, 10)
		for i := 0; i < 10; i++ {
			vals1[i] = rng1.Intn(1000)
			vals2[i] = rng2.Intn(1000)
		}

		require.NotEqual(t, vals1, vals2)
	})
}
