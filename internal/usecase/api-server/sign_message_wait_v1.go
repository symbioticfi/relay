package api_server

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// SignMessageWait handles the streaming gRPC SignMessageWait request
func (h *grpcHandler) SignMessageWait(req *apiv1.SignMessageWaitRequest, stream grpc.ServerStreamingServer[apiv1.SignMessageWaitResponse]) error {
	ctx := stream.Context()

	// First, sign the message
	sigReq := &apiv1.SignMessageRequest{
		KeyTag:        req.GetKeyTag(),
		Message:       req.GetMessage(),
		RequiredEpoch: req.RequiredEpoch,
	}
	signResp, err := h.SignMessage(ctx, sigReq)
	if err != nil {
		return err
	}

	// Send initial pending status
	err = stream.Send(&apiv1.SignMessageWaitResponse{
		Status:           apiv1.SigningStatus_SIGNING_STATUS_PENDING,
		RequestId:        signResp.GetRequestId(),
		Epoch:            signResp.GetEpoch(),
		AggregationProof: nil, // will be filled later
	})
	if err != nil {
		return err
	}

	// Poll for aggregation status and proof
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// TODO: decide timeout
	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return stream.Send(&apiv1.SignMessageWaitResponse{
				Status:           apiv1.SigningStatus_SIGNING_STATUS_TIMEOUT,
				RequestId:        signResp.GetRequestId(),
				Epoch:            signResp.GetEpoch(),
				AggregationProof: nil, // no proof yet
			})
		case <-ticker.C:
			// Check for aggregation proof
			requestID := signResp.GetRequestId()
			proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(requestID))
			if err == nil {
				// Success - send final proof
				return stream.Send(&apiv1.SignMessageWaitResponse{
					Status:    apiv1.SigningStatus_SIGNING_STATUS_COMPLETED,
					RequestId: signResp.GetRequestId(),
					Epoch:     signResp.GetEpoch(),
					AggregationProof: &apiv1.AggregationProof{
						MessageHash: proof.MessageHash,
						Proof:       proof.Proof,
					},
				})
			}
		}
	}
}
