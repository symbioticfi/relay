package apiApp

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/gen/api"
)

func (h *handler) GetAggregationProofGet(ctx context.Context, params api.GetAggregationProofGetParams) (*api.AggregationProof, error) {
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &api.AggregationProof{
		VerificationType: uint32(proof.VerificationType),
		MessageHash:      proof.MessageHash,
		Proof:            proof.Proof,
	}, nil
}
