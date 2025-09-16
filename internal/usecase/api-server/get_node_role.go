package api_server

import (
	"context"

	"github.com/go-errors/errors"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetNodeRole returns the role of the current node in the given epoch
func (h *grpcHandler) GetNodeRole(ctx context.Context, req *apiv1.GetNodeRoleRequest) (*apiv1.GetNodeRoleResponse, error) {
	latestEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}

	epochRequested := latestEpoch
	if req.Epoch != nil {
		epochRequested = req.GetEpoch()
	}

	// epoch from future
	if epochRequested > latestEpoch {
		return nil, errors.New("epoch requested is greater than latest epoch")
	}

	validatorSet, err := h.getValidatorSetForEpoch(ctx, epochRequested)
	if err != nil {
		return nil, err
	}

	privKey, err := h.cfg.KeyProvider.GetPrivateKey(validatorSet.RequiredKeyTag)
	if err != nil {
		if errors.Is(err, keyprovider.ErrKeyNotFound) {
			// key not there so no roles
			return &apiv1.GetNodeRoleResponse{}, nil
		}
		return nil, errors.Errorf("failed to get key for required key tag %s: %w", validatorSet.RequiredKeyTag, err)
	}

	return &apiv1.GetNodeRoleResponse{
		IsAggregator: validatorSet.IsAggregator(privKey.PublicKey().OnChain()),
		IsCommiter:   validatorSet.IsCommitter(privKey.PublicKey().OnChain()),
		IsSigner:     validatorSet.IsSigner(privKey.PublicKey().OnChain()),
	}, nil
}
