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

func (e *Client) multicallExists(ctx context.Context, chainId uint64) (bool, error) {
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

func (e *Client) multicall(ctx context.Context, chainId uint64, calls []Call) (_ []Result, err error) {
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
	maxCalls := e.cfg.MaxCalls
	if maxCalls != 0 {
		batches = (len(calls) + maxCalls - 1) / maxCalls
	} else {
		maxCalls = len(calls)
	}

	callOpts := &bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}

	results := make([]Result, 0, len(calls))
	for i := 0; i < batches; i++ {
		start := i * maxCalls
		end := min((i+1)*maxCalls, len(calls))
		out, err := multicall.Aggregate3(callOpts, calls[start:end])
		if err != nil {
			return nil, errors.Errorf("failed to aggregate calls: %w", err)
		}
		results = append(results, out...)
	}

	return results, nil
}

func (e *Client) getVotingPowersMulticall(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) ([]entity.OperatorVotingPower, error) {
	abi, err := gen.IVotingPowerProviderMetaData.GetAbi()
	if err != nil {
		return nil, errors.Errorf("failed to get ABI: %v", err)
	}

	operators, err := e.GetOperators(ctx, address, timestamp)
	if err != nil {
		return nil, errors.Errorf("get operators failed: %v", err)
	}

	votingPowers := make([]entity.OperatorVotingPower, 0, len(operators))

	calls := make([]Call, 0, len(operators))

	for _, operator := range operators {
		bytes, err := abi.Pack("getOperatorVotingPowersAt", operator, []byte{}, big.NewInt(int64(timestamp)))
		if err != nil {
			return nil, errors.Errorf("failed to get bytes: %v", err)
		}
		calls = append(calls, Call{
			Target:       address.Address,
			CallData:     bytes,
			AllowFailure: false,
		})
	}

	outs, err := e.multicall(ctx, address.ChainId, calls)
	if err != nil {
		return nil, errors.Errorf("multicall failed: %v", err)
	}

	if len(outs) != len(calls) {
		return nil, errors.Errorf("multicall failed: expected %d calls, got %d", len(calls), len(outs))
	}

	for i, out := range outs {
		var res []gen.IVotingPowerProviderVaultValue

		if err := abi.UnpackIntoInterface(&res, "getOperatorVotingPowersAt", out.ReturnData); err != nil {
			return nil, errors.Errorf("failed to unpack getOperatorVotingPowers: %v", err)
		}

		votingPowers = append(votingPowers, entity.OperatorVotingPower{
			Operator: operators[i],
			Vaults: lo.Map(res, func(v gen.IVotingPowerProviderVaultValue, _ int) entity.VaultVotingPower {
				return entity.VaultVotingPower{
					Vault:       v.Vault,
					VotingPower: entity.ToVotingPower(v.Value),
				}
			}),
		})
	}

	return votingPowers, nil
}

func (e *Client) getKeysMulticall(ctx context.Context, address entity.CrossChainAddress, timestamp uint64) (_ []entity.OperatorWithKeys, err error) {
	abi, err := gen.IKeyRegistryMetaData.GetAbi()
	if err != nil {
		return nil, errors.Errorf("failed to get ABI: %v", err)
	}

	operators, err := e.GetKeysOperators(ctx, address, timestamp)
	if err != nil {
		return nil, errors.Errorf("get keys operators failed: %v", err)
	}

	keys := make([]entity.OperatorWithKeys, 0, len(operators))
	calls := make([]Call, 0, len(operators))

	for _, operator := range operators {
		bytes, err := abi.Pack("getKeysAt0", operator, big.NewInt(int64(timestamp)))
		if err != nil {
			return nil, errors.Errorf("failed to get bytes: %v", err)
		}

		calls = append(calls, Call{
			Target:       address.Address,
			CallData:     bytes,
			AllowFailure: false,
		})
	}

	outs, err := e.multicall(ctx, address.ChainId, calls)
	if err != nil {
		return nil, errors.Errorf("multicall failed: %v", err)
	}

	if len(outs) != len(calls) {
		return nil, errors.Errorf("multicall failed: expected %d calls, got %d", len(calls), len(outs))
	}

	for i, out := range outs {
		var res []gen.IKeyRegistryKey

		if err := abi.UnpackIntoInterface(&res, "getKeysAt0", out.ReturnData); err != nil {
			return nil, errors.Errorf("failed to unpack getKeysAt0: %v", err)
		}

		keys = append(keys, entity.OperatorWithKeys{
			Operator: operators[i],
			Keys: lo.Map(res, func(v gen.IKeyRegistryKey, _ int) entity.ValidatorKey {
				return entity.ValidatorKey{
					Tag:     entity.KeyTag(v.Tag),
					Payload: v.Payload,
				}
			}),
		})
	}

	return keys, nil
}
