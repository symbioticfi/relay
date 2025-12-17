package evm

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) RegisterOperator(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	registry, err := e.getOperatorRegistryContract(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get operator registry contract: %w", err)
	}

	return e.doTransaction(ctx, "RegisterOperator", addr, registry.RegisterOperator)
}

func (e *Client) RegisterKey(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
	keyTag symbiotic.KeyTag,
	key symbiotic.CompactPublicKey,
	signature symbiotic.RawSignature,
	extraData []byte,
) (_ symbiotic.TxResult, err error) {
	registry, err := e.getKeyRegistryContract(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get key registry contract: %w", err)
	}

	return e.doTransaction(ctx, "SetKey", addr, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return registry.SetKey(txOpts, uint8(keyTag), key, signature, extraData)
	})
}

func (e *Client) InvalidateOldSignatures(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContractTransactor(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "InvalidateOldSignatures", addr, votingPowerProvider.InvalidateOldSignatures)
}

func (e *Client) RegisterOperatorVotingPowerProvider(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContractTransactor(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "RegisterOperatorVotingPowerProvider", addr, votingPowerProvider.RegisterOperator)
}

func (e *Client) UnregisterOperatorVotingPowerProvider(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContractTransactor(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "UnregisterOperatorVotingPowerProvider", addr, votingPowerProvider.UnregisterOperator)
}
