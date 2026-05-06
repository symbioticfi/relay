package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	internalentity "github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

// GetSignaturesByEpoch handles the gRPC GetSignaturesByEpoch request.
// Returns one page of signatures (cursor-paginated; cursor is opaque base64).
func (h *grpcHandler) GetSignaturesByEpoch(ctx context.Context, req *apiv1.GetSignaturesByEpochRequest) (*apiv1.GetSignaturesByEpochResponse, error) {
	pageSize := clampPageSize(int(req.GetPageSize()), defaultListPageSize, maxListPageSize)

	from, err := decodeCursor(req.GetFrom())
	if err != nil {
		return nil, err
	}

	signatures, next, err := h.cfg.Repo.GetSignaturesByEpoch(ctx, entity.Epoch(req.GetEpoch()), pageSize, from)
	if err != nil {
		if errors.Is(err, internalentity.ErrInvalidCursor) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, errors.Errorf("failed to get signatures by epoch: %w", err)
	}

	return &apiv1.GetSignaturesByEpochResponse{
		Signatures: convertSignaturesToPB(signatures),
		NextFrom:   encodeCursor(next),
	}, nil
}
