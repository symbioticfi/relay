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

func (h *grpcHandler) ListenSignatures(
	req *apiv1.ListenSignaturesRequest,
	stream grpc.ServerStreamingServer[apiv1.ListenSignaturesResponse],
) error {
	ctx := stream.Context()

	if h.signatureHub.Count() >= h.cfg.MaxAllowedStreamsCount {
		return status.Errorf(codes.ResourceExhausted, "too many signatures")
	}

	subscriptionID := uuid.New()

	signatureCh := h.signatureHub.Subscribe(subscriptionID.String())
	defer h.signatureHub.Unsubscribe(subscriptionID.String())

	if epoch := req.GetStartEpoch(); epoch != 0 {
		latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
		if err != nil {
			if !errors.Is(err, entity.ErrEntityNotFound) {
				return err
			}
			latestEpoch = 0
		}

		for e := symbiotic.Epoch(epoch); e <= latestEpoch; e++ {
			var from []byte
			for {
				signatures, next, err := h.cfg.Repo.GetSignaturesByEpoch(ctx, e, maxListPageSize, from)
				if err != nil {
					return err
				}

				for _, signature := range signatures {
					if err = stream.Send(&apiv1.ListenSignaturesResponse{
						RequestId: signature.RequestID().Hex(),
						Epoch:     uint64(signature.Epoch),
						Signature: convertSignatureToPB(signature),
					}); err != nil {
						return err
					}
				}

				if next == nil {
					break
				}
				from = next
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case signature := <-signatureCh:
			if err := stream.Send(&apiv1.ListenSignaturesResponse{
				RequestId: signature.RequestID().Hex(),
				Epoch:     uint64(signature.Epoch),
				Signature: convertSignatureToPB(signature),
			}); err != nil {
				return err
			}
		}
	}
}
