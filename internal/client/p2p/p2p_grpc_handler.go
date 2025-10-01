package p2p

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	p2pv1 "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/pkg/log"
)

// syncRequestHandler defines the interface for handling both signature and aggregation proof requests
type syncRequestHandler interface {
	HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
	HandleWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error)
}

type GRPCHandler struct {
	p2pv1.UnimplementedSymbioticP2PServiceServer

	syncHandler syncRequestHandler
}

func NewP2PHandler(syncHandler syncRequestHandler) *GRPCHandler {
	return &GRPCHandler{
		UnimplementedSymbioticP2PServiceServer: p2pv1.UnimplementedSymbioticP2PServiceServer{},
		syncHandler:                            syncHandler,
	}
}

// WantSignatures handles incoming signature requests from peers
func (h *GRPCHandler) WantSignatures(ctx context.Context, req *p2pv1.WantSignaturesRequest) (*p2pv1.WantSignaturesResponse, error) {
	ctx = log.WithComponent(ctx, "p2p-grpc-handler")

	entityReq, err := protoToEntityRequest(req)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to convert request: %w", err)
	}

	response, err := h.syncHandler.HandleWantSignaturesRequest(ctx, entityReq)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to handle request: %w", err)
	}

	return entityToProtoResponse(response), nil
}

// WantAggregationProofs handles incoming aggregation proof requests from peers
func (h *GRPCHandler) WantAggregationProofs(ctx context.Context, req *p2pv1.WantAggregationProofsRequest) (*p2pv1.WantAggregationProofsResponse, error) {
	ctx = log.WithComponent(ctx, "p2p-grpc-handler")

	entityReq := protoToEntityAggregationProofRequest(req)

	response, err := h.syncHandler.HandleWantAggregationProofsRequest(ctx, entityReq)
	if err != nil {
		return &p2pv1.WantAggregationProofsResponse{}, errors.Errorf("failed to handle aggregation proof request: %w", err)
	}

	return entityToProtoAggregationProofResponse(response), nil
}
