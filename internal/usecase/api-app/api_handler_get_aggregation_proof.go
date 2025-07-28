package apiApp

import (
	"context"

	api2 "github.com/symbioticfi/relay/core/api/gen/api"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

func (h *handler) GetAggregationProofGet(ctx context.Context, params api2.GetAggregationProofGetParams) (*api2.AggregationProof, error) {
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &api2.AggregationProof{
		VerificationType: uint32(proof.VerificationType),
		MessageHash:      proof.MessageHash,
		Proof:            proof.Proof,
	}, nil
}
