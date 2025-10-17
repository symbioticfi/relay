package api_server

import (
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (h *grpcHandler) ListenValidatorSet(
	req *apiv1.ListenValidatorSetRequest,
	stream grpc.ServerStreamingServer[apiv1.ListenValidatorSetResponse],
) error {
	ctx := stream.Context()

	if h.validatorSetsHub.Count() >= h.cfg.MaxAllowedStreamsCount {
		return status.Errorf(codes.ResourceExhausted, "max allowed streams limit reached")
	}

	subscriptionID := uuid.New()

	validatorSetCh := h.validatorSetsHub.Subscribe(subscriptionID.String())
	defer h.validatorSetsHub.Unsubscribe(subscriptionID.String())

	if epoch := req.GetStartEpoch(); epoch != 0 {
		validatorSets, err := h.cfg.Repo.GetValidatorSetsByEpoch(ctx, symbiotic.Epoch(epoch))
		if err != nil {
			return err
		}

		for _, valSet := range validatorSets {
			if err = stream.Send(convertValidatorSetToStreamResponse(valSet)); err != nil {
				return err
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case valSet := <-validatorSetCh:
			if err := stream.Send(convertValidatorSetToStreamResponse(valSet)); err != nil {
				return err
			}
		}
	}
}

// convertValidatorSetToStreamResponse converts ValidatorSet to ListenValidatorSetResponse
func convertValidatorSetToStreamResponse(valSet symbiotic.ValidatorSet) *apiv1.ListenValidatorSetResponse {
	return &apiv1.ListenValidatorSetResponse{
		Version:          uint32(valSet.Version),
		RequiredKeyTag:   uint32(valSet.RequiredKeyTag),
		Epoch:            uint64(valSet.Epoch),
		CaptureTimestamp: timestamppb.New(time.Unix(int64(valSet.CaptureTimestamp), 0).UTC()),
		QuorumThreshold:  valSet.QuorumThreshold.String(),
		Status:           convertValidatorSetStatusToPB(valSet.Status),
		Validators: lo.Map(valSet.Validators, func(v symbiotic.Validator, _ int) *apiv1.Validator {
			return convertValidatorToPB(v)
		}),
	}
}
