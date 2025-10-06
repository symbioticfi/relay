package api_server

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"

	"github.com/symbioticfi/relay/core/entity"
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

	requestID := signResp.GetRequestId()
	requestIDHash := common.HexToHash(requestID)

	// Check if proof already exists (early optimization)
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, requestIDHash)
	if err == nil {
		// Proof already exists, return immediately
		return stream.Send(&apiv1.SignMessageWaitResponse{
			Status:    apiv1.SigningStatus_SIGNING_STATUS_COMPLETED,
			RequestId: requestID,
			Epoch:     signResp.GetEpoch(),
			AggregationProof: &apiv1.AggregationProof{
				MessageHash: proof.MessageHash,
				Proof:       proof.Proof,
			},
		})
	}

	// Send initial pending status
	err = stream.Send(&apiv1.SignMessageWaitResponse{
		Status:           apiv1.SigningStatus_SIGNING_STATUS_PENDING,
		RequestId:        requestID,
		Epoch:            signResp.GetEpoch(),
		AggregationProof: nil, // will be filled later
	})
	if err != nil {
		return err
	}

	// Create subscription channel for this request
	proofCh := make(chan entity.AggregationProof, 1)
	h.subscriptions.Store(requestID, proofCh)
	defer func() {
		h.subscriptions.Delete(requestID)
		close(proofCh)
	}()

	// TODO: decide timeout
	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	// Wait for proof via signal or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timeout.C:
		return stream.Send(&apiv1.SignMessageWaitResponse{
			Status:           apiv1.SigningStatus_SIGNING_STATUS_TIMEOUT,
			RequestId:        requestID,
			Epoch:            signResp.GetEpoch(),
			AggregationProof: nil, // no proof yet
		})
	case proofFromCh := <-proofCh:
		// Success - received proof from signal
		return stream.Send(&apiv1.SignMessageWaitResponse{
			Status:    apiv1.SigningStatus_SIGNING_STATUS_COMPLETED,
			RequestId: requestID,
			Epoch:     signResp.GetEpoch(),
			AggregationProof: &apiv1.AggregationProof{
				MessageHash: proofFromCh.MessageHash,
				Proof:       proofFromCh.Proof,
			},
		})
	}
}
