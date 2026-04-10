package api_server

import (
	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/symbioticfi/relay/internal/entity"
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
		latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
		if err != nil {
			if !errors.Is(err, entity.ErrEntityNotFound) {
				return err
			}
			latestEpoch = 0
		}

		for e := symbiotic.Epoch(epoch); e <= latestEpoch; e++ {
			valSet, err := h.cfg.Repo.GetValidatorSetByEpoch(ctx, e)
			if err != nil {
				if errors.Is(err, entity.ErrEntityNotFound) {
					continue
				}
				return err
			}

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
