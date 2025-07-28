package apiApp

import (
	"context"
	"time"

	"github.com/symbioticfi/relay/core/api/gen/api"

	"github.com/go-errors/errors"
)

func (h *handler) GetCurrentEpochGet(ctx context.Context) (*api.GetCurrentEpochGetOK, error) {
	currentEpoch, err := h.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get current epoch: %w", err)
	}

	epochStart, err := h.cfg.EvmClient.GetEpochStart(ctx, currentEpoch)
	if err != nil {
		return nil, errors.Errorf("failed to get epoch start: %w", err)
	}

	return &api.GetCurrentEpochGetOK{
		Epoch:     currentEpoch,
		StartTime: time.Unix(int64(epochStart), 0).UTC(),
	}, nil
}
