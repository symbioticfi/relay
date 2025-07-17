package apiApp

import (
	"context"
	"time"

	"github.com/go-errors/errors"

	"github.com/symbiotic/relay/internal/gen/api"
)

func (h *handler) GetSuggestedEpochGet(ctx context.Context) (*api.GetSuggestedEpochGetOK, error) {
	valset, err := h.cfg.Repo.GetLatestValidatorSet(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest validator set: %w", err)
	}

	// just return latest derived epoch
	// there is no way to make it more deterministic across each node
	return &api.GetSuggestedEpochGetOK{
		Epoch:     valset.Epoch,
		StartTime: time.Unix(int64(valset.CaptureTimestamp), 0).UTC(),
	}, nil
}
