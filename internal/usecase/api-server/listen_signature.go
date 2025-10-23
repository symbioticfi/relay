package api_server

import (
	"github.com/google/uuid"
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
		signatures, err := h.cfg.Repo.GetSignaturesStartingFromEpoch(ctx, symbiotic.Epoch(epoch))
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
