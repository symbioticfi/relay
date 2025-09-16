package api_server

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetValidatorSetHeader handles the gRPC GetValidatorSetHeader request
func (h *grpcHandler) GetValidatorSetHeader(ctx context.Context, req *apiv1.GetValidatorSetHeaderRequest) (*apiv1.GetValidatorSetHeaderResponse, error) {
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

	// get header from validator set
	header, err := validatorSet.GetHeader()
	if err != nil {
		return nil, errors.Errorf("failed to get validator set header: %w", err)
	}

	return &apiv1.GetValidatorSetHeaderResponse{
		Version:            uint32(header.Version),
		RequiredKeyTag:     uint32(header.RequiredKeyTag),
		Epoch:              header.Epoch,
		CaptureTimestamp:   timestamppb.New(time.Unix(int64(header.CaptureTimestamp), 0).UTC()),
		QuorumThreshold:    header.QuorumThreshold.String(),
		TotalVotingPower:   header.TotalVotingPower.String(),
		ValidatorsSszMroot: header.ValidatorsSszMRoot.Hex(),
	}, nil
}
