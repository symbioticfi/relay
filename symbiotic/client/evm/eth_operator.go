package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) RegisterOperator(
	ctx context.Context,
	opRegistryAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	registry, err := e.getOperatorRegistryContract(opRegistryAddr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get operator registry contract: %w", err)
	}

	return e.doTransaction(ctx, "RegisterOperator", e.driverChainID, registry.RegisterOperator)
}

func (e *Client) RegisterKey(
	ctx context.Context,
	keyRegistryAddr symbiotic.CrossChainAddress,
	keyTag symbiotic.KeyTag,
	key symbiotic.CompactPublicKey,
	signature symbiotic.RawSignature,
	extraData []byte,
) (_ symbiotic.TxResult, err error) {
	registry, err := e.getKeyRegistryContract(keyRegistryAddr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get key registry contract: %w", err)
	}

	return e.doTransaction(ctx, "SetKey", keyRegistryAddr.ChainId, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return registry.SetKey(txOpts, uint8(keyTag), key, signature, extraData)
	})
}

func (e *Client) InvalidateOldSignatures(
	ctx context.Context,
	votingPowerProviderAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContract(votingPowerProviderAddr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "InvalidateOldSignatures", e.driverChainID, votingPowerProvider.InvalidateOldSignatures)
}

func (e *Client) RegisterOperatorVotingPowerProvider(
	ctx context.Context,
	votingPowerProviderAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContract(votingPowerProviderAddr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "RegisterOperatorVotingPowerProvider", e.driverChainID, votingPowerProvider.RegisterOperator)
}

func (e *Client) IsOperatorRegistered(
	ctx context.Context,
	votingPowerProviderAddr, operator symbiotic.CrossChainAddress,
) (_ bool, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("IsOperatorRegistered", e.driverChainID, err, now)
	}(time.Now())

	votingPowerProvider, err := e.getVotingPowerProviderContract(votingPowerProviderAddr)
	if err != nil {
		return false, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return votingPowerProvider.IsOperatorRegistered(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, operator.Address)
}

func (e *Client) UnregisterOperatorVotingPowerProvider(
	ctx context.Context,
	votingPowerProviderAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	votingPowerProvider, err := e.getVotingPowerProviderContract(votingPowerProviderAddr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	return e.doTransaction(ctx, "UnregisterOperatorVotingPowerProvider", e.driverChainID, votingPowerProvider.UnregisterOperator)
}
