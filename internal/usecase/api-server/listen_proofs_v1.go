package api_server

import (
	"github.com/google/uuid"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *grpcHandler) ListenProofs(
	req *apiv1.ListenProofsRequest,
	stream grpc.ServerStreamingServer[apiv1.ListenProofsResponse],
) error {
	ctx := stream.Context()

	if h.proofsHub.Count() >= h.cfg.MaxAllowedStreamsCount {
		return status.Errorf(codes.ResourceExhausted, "max allowed streams limit reached")
	}

	subscriptionID := uuid.New()

	proofsCh := h.proofsHub.Subscribe(subscriptionID.String())
	defer h.proofsHub.Unsubscribe(subscriptionID.String())

	if epoch := req.GetStartEpoch(); epoch != 0 {
		proofs, err := h.cfg.Repo.GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(epoch))
		if err != nil {
			return err
		}

		for _, proof := range proofs {
			if err = stream.Send(&apiv1.ListenProofsResponse{
				RequestId: proof.RequestID().Hex(),
				Epoch:     uint64(proof.Epoch),
				AggregationProof: &apiv1.AggregationProof{
					MessageHash: proof.MessageHash,
					Proof:       proof.Proof,
					RequestId:   proof.RequestID().Hex(),
				},
			}); err != nil {
				return err
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case proof := <-proofsCh:
			if err := stream.Send(&apiv1.ListenProofsResponse{
				RequestId: proof.RequestID().Hex(),
				Epoch:     uint64(proof.Epoch),
				AggregationProof: &apiv1.AggregationProof{
					MessageHash: proof.MessageHash,
					Proof:       proof.Proof,
					RequestId:   proof.RequestID().Hex(),
				},
			}); err != nil {
				return err
			}
		}
	}
}
