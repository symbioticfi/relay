package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetSignatures handles the gRPC GetSignatures request
func (h *grpcHandler) GetSignatures(ctx context.Context, req *v1.GetSignaturesRequest) (*v1.GetSignaturesResponse, error) {
	signatures, err := h.cfg.Repo.GetAllSignatures(ctx, common.HexToHash(req.GetRequestHash()))
	if err != nil {
		return nil, errors.Errorf("failed to get signatures: %w", err)
	}

	return &v1.GetSignaturesResponse{
		Signatures: convertSignaturesToPB(signatures),
	}, nil
}

func convertSignaturesToPB(signatures []entity.SignatureExtended) []*v1.Signature {
	return lo.Map(signatures, func(sig entity.SignatureExtended, _ int) *v1.Signature {
		return convertSignatureToPB(sig)
	})
}

func convertSignatureToPB(sig entity.SignatureExtended) *v1.Signature {
	return &v1.Signature{
		Signature:   sig.Signature,
		MessageHash: sig.MessageHash,
		PublicKey:   sig.PublicKey,
	}
}
