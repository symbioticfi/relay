package strategyHelpers

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
)

func GetPreviousHashForCommittedValset(ctx context.Context, client evm.IEvmClient, committedAddr entity.CrossChainAddress, epoch uint64, valset entity.ValidatorSet) (common.Hash, error) {
	previousHeaderHash, err := client.GetPreviousHeaderHashAt(ctx, committedAddr, epoch)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get previous header hash: %w", err)
	}
	// valset integrity check
	valset.PreviousHeaderHash = previousHeaderHash
	committedHash, err := client.GetHeaderHashAt(ctx, committedAddr, epoch)
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
	}
	valsetHeader, err := valset.GetHeader()
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
	}
	calculatedHash, err := valsetHeader.Hash()
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
	}

	if !bytes.Equal(committedHash[:], calculatedHash[:]) {
		slog.DebugContext(ctx, "Validator set integrity check failed", "committed hash", committedHash, "calculated hash", calculatedHash)
		return common.Hash{}, errors.Errorf("validator set hash mistmach at epoch %d", epoch)
	}
	slog.DebugContext(ctx, "Validator set integrity check passed", "hash", committedHash)

	return previousHeaderHash, nil
}
