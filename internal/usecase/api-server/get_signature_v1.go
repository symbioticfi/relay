package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetSignatures handles the gRPC GetSignatures request
func (h *grpcHandler) GetSignatures(ctx context.Context, req *apiv1.GetSignaturesRequest) (*apiv1.GetSignaturesResponse, error) {
	signatures, err := h.cfg.Repo.GetAllSignatures(ctx, common.HexToHash(req.GetRequestId()))
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "signatures for request %s not found", req.GetRequestId())
		}
		return nil, errors.Errorf("failed to get signatures: %w", err)
	}

	return &apiv1.GetSignaturesResponse{
		Signatures: convertSignaturesToPB(signatures),
	}, nil
}

func convertSignaturesToPB(signatures []symbiotic.Signature) []*apiv1.Signature {
	return lo.Map(signatures, func(sig symbiotic.Signature, _ int) *apiv1.Signature {
		return convertSignatureToPB(sig)
	})
}

func convertSignatureToPB(sig symbiotic.Signature) *apiv1.Signature {
	return &apiv1.Signature{
		Signature:   sig.Signature,
		MessageHash: sig.MessageHash,
		PublicKey:   sig.PublicKey.Raw(),
	}
}
