package api_server

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetCustomScheduleNodeStatus handles the gRPC GetCustomScheduleNodeStatus request
func (h *grpcHandler) GetCustomScheduleNodeStatus(ctx context.Context, req *apiv1.GetCustomScheduleNodeStatusRequest) (*apiv1.GetCustomScheduleNodeStatusResponse, error) {
	// Validate request parameters
	if req.GetSlotDurationSeconds() == 0 {
		return nil, status.Error(codes.InvalidArgument, "slot_duration_seconds must be greater than 0")
	}
	if req.GetMaxParticipantsPerSlot() == 0 {
		return nil, status.Error(codes.InvalidArgument, "max_participants_per_slot must be greater than 0")
	}
	if req.GetMinParticipantsPerSlot() > req.GetMaxParticipantsPerSlot() {
		return nil, status.Error(codes.InvalidArgument, "min_participants_per_slot cannot be greater than max_participants_per_slot")
	}

	// Get the latest epoch if not provided
	latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set epoch: %w", err)
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = symbiotic.Epoch(req.GetEpoch())
	}

	// Validate epoch is not from the future
	if epochRequested > latestEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "epoch %d is greater than latest epoch %d", epochRequested, latestEpoch)
	}

	// Get validator set for the epoch
	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	// Get epoch start time
	epochStartTime := time.Unix(int64(validatorSet.CaptureTimestamp), 0)

	// Get local validator address
	pubkey, err := h.cfg.KeyProvider.GetOnchainKeyFromCache(validatorSet.RequiredKeyTag)
	if err != nil {
		return nil, errors.Errorf("failed to get onchain key from cache: %w", err)
	}

	activeValidators := validatorSet.Validators.GetActiveValidators()
	if len(activeValidators) == 0 {
		return nil, status.Error(codes.Internal, "no active validators found in validator set")
	}

	currentSlot, slotStartTime, err := getCurrentSlot(
		epochStartTime,
		time.Now,
		req.GetSlotDurationSeconds(),
	)
	if err != nil {
		return nil, err
	}

	slotEndTime := slotStartTime.Add(time.Duration(req.GetSlotDurationSeconds()) * time.Second)

	// Dev: currently only returns schedule for active validators so returning false if not active in valset
	val, ok := activeValidators.FindValidatorByKey(validatorSet.RequiredKeyTag, pubkey)
	if !ok {
		return &apiv1.GetCustomScheduleNodeStatusResponse{
			IsActive:             false,
			CurrentSlotEndTime:   timestamppb.New(slotEndTime),
			CurrentSlotStartTime: timestamppb.New(slotStartTime),
		}, nil
	}

	// Check if node is active in the custom schedule
	isActive, err := isNodeActiveInCustomSchedule(
		activeValidators,
		epochRequested,
		req.GetSeed(),
		currentSlot,
		req.GetMaxParticipantsPerSlot(),
		req.GetMinParticipantsPerSlot(),
		val.Operator,
	)
	if err != nil {
		return nil, err
	}

	return &apiv1.GetCustomScheduleNodeStatusResponse{
		IsActive:             isActive,
		CurrentSlotEndTime:   timestamppb.New(slotEndTime),
		CurrentSlotStartTime: timestamppb.New(slotStartTime),
	}, nil
}

// getCurrentSlot calculates the current slot number and its start time based on epoch start time and slot duration
// Returns: currentSlot, slotStartTime, error
func getCurrentSlot(
	epochStartTime time.Time,
	currentTimeFunc func() time.Time,
	slotDurationSeconds uint64,
) (uint64, time.Time, error) {
	// Calculate current slot number
	now := currentTimeFunc()
	elapsedTime := now.Sub(epochStartTime)
	if elapsedTime < 0 {
		// Current time is before epoch start
		return 0, time.Time{}, status.Error(codes.InvalidArgument, "epoch has not started yet")
	}

	currentSlot := uint64(elapsedTime.Seconds()) / slotDurationSeconds

	// Calculate slot start and end times
	return currentSlot, epochStartTime.Add(time.Duration(currentSlot*slotDurationSeconds) * time.Second), nil
}

// isNodeActiveInCustomSchedule checks if the node is active at the current time in a custom schedule
// Returns: isActive, slotStartTime, slotEndTime, error
func isNodeActiveInCustomSchedule(
	activeValidators symbiotic.Validators,
	epoch symbiotic.Epoch,
	seed []byte,
	currentSlot uint64,
	maxParticipantsPerSlot uint32,
	minParticipantsPerSlot uint32,
	localAddress common.Address,
) (bool, error) {
	// Create a deterministic random number generator seeded with epoch and seed
	rng := createSeededRNG(epoch, seed)

	// Create shuffled indices array
	indices := make([]int, len(activeValidators))
	for i := range indices {
		indices[i] = i
	}

	// Fisher-Yates shuffle
	rng.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Calculate number of groups based on max and min constraints
	maxGroups := len(activeValidators) / int(maxParticipantsPerSlot)
	remainder := len(activeValidators) % int(maxParticipantsPerSlot)
	if remainder != 0 && remainder >= int(minParticipantsPerSlot) {
		maxGroups++
	}

	// If no valid groups can be formed, no validator is active
	if maxGroups == 0 {
		return false, status.Errorf(codes.InvalidArgument, "no valid groups can be formed with the given parameters. Total validators=%d, maxParticipantsPerSlot=%d, minParticipantsPerSlot=%d", len(activeValidators), maxParticipantsPerSlot, minParticipantsPerSlot)
	}

	// Groups cycle through slots using modulo
	groupIdx := currentSlot % uint64(maxGroups)

	// Check if local validator is in the current group
	startIdx := groupIdx * uint64(maxParticipantsPerSlot)
	endIdx := min((groupIdx+1)*uint64(maxParticipantsPerSlot), uint64(len(indices)))

	for i := startIdx; i < endIdx; i++ {
		if activeValidators[indices[i]].Operator == localAddress {
			return true, nil
		}
	}

	return false, nil
}

// createSeededRNG creates a deterministic random number generator seeded with epoch and seed bytes
func createSeededRNG(epoch symbiotic.Epoch, seed []byte) *rand.Rand {
	// Create a deterministic seed by combining epoch and seed
	hasher := sha256.New()
	hasher.Write(epoch.Bytes())
	if len(seed) > 0 {
		hasher.Write(seed)
	}

	// Get hash and convert to int64 seed
	hash := hasher.Sum(nil)
	seedValue := int64(binary.BigEndian.Uint64(hash[:8]))

	// Create new random source and generator
	source := rand.NewSource(seedValue)
	return rand.New(source) //nolint:gosec // we need deterministic randomness here so need to use math/rand
}
