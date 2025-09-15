package p2p

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	p2pv1 "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/pkg/log"
)

// signatureRequestHandler defines the interface for handling signature requests
type signatureRequestHandler interface {
	HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

// aggregationProofRequestHandler defines the interface for handling aggregation proof requests
type aggregationProofRequestHandler interface {
	HandleWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error)
}

type GRPCHandler struct {
	p2pv1.UnimplementedSymbioticP2PServiceServer

	signatureHandler        signatureRequestHandler
	aggregationProofHandler aggregationProofRequestHandler
}

func NewP2PHandler(signatureHandler signatureRequestHandler, aggregationProofHandler aggregationProofRequestHandler) *GRPCHandler {
	return &GRPCHandler{
		signatureHandler:        signatureHandler,
		aggregationProofHandler: aggregationProofHandler,
	}
}

// WantSignatures handles incoming signature requests from peers
func (h *GRPCHandler) WantSignatures(ctx context.Context, req *p2pv1.WantSignaturesRequest) (*p2pv1.WantSignaturesResponse, error) {
	ctx = log.WithComponent(ctx, "p2p-grpc-handler")

	entityReq, err := protoToEntityRequest(req)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to convert request: %w", err)
	}

	response, err := h.signatureHandler.HandleWantSignaturesRequest(ctx, entityReq)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to handle request: %w", err)
	}

	return entityToProtoResponse(response), nil
}

// WantAggregationProofs handles incoming aggregation proof requests from peers
func (h *GRPCHandler) WantAggregationProofs(ctx context.Context, req *p2pv1.WantAggregationProofsRequest) (*p2pv1.WantAggregationProofsResponse, error) {
	ctx = log.WithComponent(ctx, "p2p-grpc-handler")

	entityReq, err := protoToEntityAggregationProofRequest(req)
	if err != nil {
		return &p2pv1.WantAggregationProofsResponse{}, errors.Errorf("failed to convert aggregation proof request: %w", err)
	}

	response, err := h.aggregationProofHandler.HandleWantAggregationProofsRequest(ctx, entityReq)
	if err != nil {
		return &p2pv1.WantAggregationProofsResponse{}, errors.Errorf("failed to handle aggregation proof request: %w", err)
	}

	return entityToProtoAggregationProofResponse(response), nil
}
