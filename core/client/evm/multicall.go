package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/symbioticfi/relay/core/client/evm/gen"
	"github.com/symbioticfi/relay/core/entity"
)

const Multicall3 = "0xcA11bde05977b3631167028862bE2a173976CA11"

type Call = gen.Multicall3Call3
type Result = gen.Multicall3Result

func (e *Client) MulticallExists(ctx context.Context, chainId uint64) (bool, error) {
	client, ok := e.conns[chainId]
	if !ok {
		return false, errors.Errorf("no connection for chain ID %d: %w", chainId, entity.ErrChainNotFound)
	}

	code, err := client.CodeAt(ctx, common.HexToAddress(Multicall3), new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()))
	if err != nil {
		return false, errors.Errorf("failed to get Multicall3 code: %w", err)
	}

	return len(code) > 0, nil
}

func (e *Client) GetOperatorVotingPowersCall(target entity.CrossChainAddress, operator common.Address, timestamp uint64) (Call, error) {
	abi, err := gen.IVotingPowerProviderMetaData.GetAbi()
	if err != nil {
		return Call{}, errors.Errorf("failed to get ABI: %v", err)
	}

	bytes, err := abi.Pack("getOperatorVotingPowersAt", operator, []byte{}, big.NewInt(int64(timestamp)))
	return Call{
		Target:       target.Address,
		CallData:     bytes,
		AllowFailure: false,
	}, err
}

func (e *Client) UnpackGetOperatorVotingPowersCall(out []byte) ([]entity.VaultVotingPower, error) {
	var res []gen.IVotingPowerProviderVaultVotingPower

	abi, err := gen.IVotingPowerProviderMetaData.GetAbi()
	if err != nil {
		return nil, errors.Errorf("failed to get ABI: %v", err)
	}

	if err := abi.UnpackIntoInterface(&res, "getOperatorVotingPowersAt", out); err != nil {
		return nil, errors.Errorf("failed to unpack getOperatorVotingPowers: %v", err)
	}

	return lo.Map(res, func(v gen.IVotingPowerProviderVaultVotingPower, _ int) entity.VaultVotingPower {
		return entity.VaultVotingPower{
			Vault:       v.Vault,
			VotingPower: entity.ToVotingPower(v.VotingPower),
		}
	}), nil
}

func (e *Client) Multicall(ctx context.Context, chainId uint64, calls []Call) (_ []Result, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("Multicall", err, now)
	}(time.Now())

	client, ok := e.conns[chainId]
	if !ok {
		return nil, errors.Errorf("no connection for chain ID %d: %w", chainId, entity.ErrChainNotFound)
	}

	multicall, err := gen.NewMulticall3(common.HexToAddress(Multicall3), client)
	if err != nil {
		return nil, errors.Errorf("failed to create multicall3: %v", err)
	}

	batches := 1
	if e.cfg.MaxCalls != 0 {
		batches = (len(calls) + e.cfg.MaxCalls - 1) / e.cfg.MaxCalls
	}

	callOpts := &bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}

	if batches == 1 {
		out, err := multicall.Aggregate3(callOpts, calls)
		if err != nil {
			return nil, errors.Errorf("failed to aggregate calls: %w", err)
		}

		return out, nil
	}

	results := make([]Result, 0, len(calls))
	for i := 0; i < batches; i++ {
		start := i * e.cfg.MaxCalls
		end := min((i+1)*e.cfg.MaxCalls, len(calls))
		out, err := multicall.Aggregate3(callOpts, calls[start:end])
		if err != nil {
			return nil, errors.Errorf("failed to aggregate calls: %w", err)
		}
		results = append(results, out...)
	}

	return results, nil
}
