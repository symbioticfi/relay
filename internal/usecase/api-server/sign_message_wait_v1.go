package api_server

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/symbioticfi/relay/internal/gen/api/v1"
	"google.golang.org/grpc"
)

// SignMessageWait handles the streaming gRPC SignMessageWait request
func (h *grpcHandler) SignMessageWait(req *v1.SignMessageWaitRequest, stream grpc.ServerStreamingServer[v1.SignMessageWaitResponse]) error {
	ctx := stream.Context()

	// First, sign the message
	sigReq := &v1.SignMessageRequest{
		KeyTag:        req.GetKeyTag(),
		Message:       req.GetMessage(),
		RequiredEpoch: req.RequiredEpoch,
	}
	signResp, err := h.SignMessage(ctx, sigReq)
	if err != nil {
		return err
	}

	// Send initial pending status
	err = stream.Send(&v1.SignMessageWaitResponse{
		Status:      v1.SigningStatus_SIGNING_STATUS_PENDING,
		RequestHash: signResp.GetRequestHash(),
		Epoch:       signResp.GetEpoch(),
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
			return stream.Send(&v1.SignMessageWaitResponse{
				Status:      v1.SigningStatus_SIGNING_STATUS_TIMEOUT,
				RequestHash: signResp.GetRequestHash(),
				Epoch:       signResp.GetEpoch(),
			})
		case <-ticker.C:
			// Check for aggregation proof
			reqHash := signResp.GetRequestHash()
			proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(reqHash))
			if err == nil {
				// Success - send final proof
				return stream.Send(&v1.SignMessageWaitResponse{
					Status:      v1.SigningStatus_SIGNING_STATUS_COMPLETED,
					RequestHash: signResp.GetRequestHash(),
					Epoch:       signResp.GetEpoch(),
					AggregationProof: &v1.AggregationProof{
						VerificationType: uint32(proof.VerificationType),
						MessageHash:      proof.MessageHash,
						Proof:            proof.Proof,
					},
				})
			}
		}
	}
}
