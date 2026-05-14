package api_server

import (
	"context"

	"github.com/go-errors/errors"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func (h *grpcHandler) GetSignaturesByEpoch(ctx context.Context, req *apiv1.GetSignaturesByEpochRequest) (*apiv1.GetSignaturesByEpochResponse, error) {
	pageSize := clampPageSize(int(req.GetPageSize()), defaultListPageSize, maxListPageSize)

	from, err := decodeCursor(req.GetFrom())
	if err != nil {
		return nil, err
	}

	signatures, next, err := h.cfg.Repo.GetSignaturesByEpoch(ctx, entity.Epoch(req.GetEpoch()), pageSize, from)
	if err != nil {
		if e := asCursorErr(err); e != nil {
			return nil, e
		}
		return nil, errors.Errorf("failed to get signatures by epoch: %w", err)
	}

	return &apiv1.GetSignaturesByEpochResponse{
		Signatures: convertSignaturesToPB(signatures),
		NextFrom:   encodeCursor(next),
	}, nil
}
