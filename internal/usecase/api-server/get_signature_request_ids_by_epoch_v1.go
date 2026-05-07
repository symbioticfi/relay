package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func (h *grpcHandler) GetSignatureRequestIDsByEpoch(ctx context.Context, req *apiv1.GetSignatureRequestIDsByEpochRequest) (*apiv1.GetSignatureRequestIDsByEpochResponse, error) {
	pageSize := clampPageSize(int(req.GetPageSize()), defaultIDListPageSize, maxIDListPageSize)

	from, err := decodeCursor(req.GetFrom())
	if err != nil {
		return nil, err
	}

	requestIDs, next, err := h.cfg.Repo.GetSignatureRequestIDsByEpoch(ctx, entity.Epoch(req.GetEpoch()), pageSize, from)
	if err != nil {
		if e := asCursorErr(err); e != nil {
			return nil, e
		}
		return nil, errors.Errorf("failed to get signature request IDs by epoch: %w", err)
	}

	return &apiv1.GetSignatureRequestIDsByEpochResponse{
		RequestIds: lo.Map(requestIDs, func(requestID common.Hash, _ int) string { return requestID.Hex() }),
		NextFrom:   encodeCursor(next),
	}, nil
}
