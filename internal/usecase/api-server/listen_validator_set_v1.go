package api_server

import (
	"github.com/google/uuid"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		validatorSets, err := h.cfg.Repo.GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(epoch))
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

func convertValidatorSetToStreamResponse(valSet symbiotic.ValidatorSet) *apiv1.ListenValidatorSetResponse {
	return &apiv1.ListenValidatorSetResponse{ValidatorSet: convertValidatorSetToPB(valSet)}
}
