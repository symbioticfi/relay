package evm

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/pkg/tracing"
	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type tracingDriver struct {
	base driverContract
}

type tracingConn struct {
	base    conn
	chainID uint64
}

var (
	_ conn                         = (*tracingConn)(nil)
	_ bind.PendingContractCaller   = (*tracingConn)(nil)
	_ bind.BlockHashContractCaller = (*tracingConn)(nil)
)

func newTracingConn(chainID uint64, base conn) conn {
	if base == nil {
		return nil
	}

	if _, ok := base.(*tracingConn); ok {
		return base
	}

	return &tracingConn{
		base:    base,
		chainID: chainID,
	}
}

func (t *tracingConn) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.CodeAt",
		t.spanAttributes("CodeAt",
			tracing.AttrAddress.String(contract.Hex()),
			attribute.String("block.number", blockNumberValue(blockNumber)),
		)...,
	)
	defer span.End()

	code, err := t.base.CodeAt(ctx, contract, blockNumber)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(code)),
	)

	return code, nil
}

func (t *tracingConn) CodeAtHash(ctx context.Context, contract common.Address, blockHash common.Hash) ([]byte, error) {
	ctx = ensureContext(ctx)
	blockHasher, ok := t.base.(bind.BlockHashContractCaller)
	if !ok {
		return nil, bind.ErrNoBlockHashState
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.CodeAtHash",
		t.spanAttributes("CodeAtHash",
			tracing.AttrAddress.String(contract.Hex()),
			attribute.String("block.hash", blockHash.Hex()),
		)...,
	)
	defer span.End()

	code, err := blockHasher.CodeAtHash(ctx, contract, blockHash)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(code)),
	)

	return code, nil
}

func (t *tracingConn) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ctx = ensureContext(ctx)
	attrs := append(t.spanAttributes("CallContract",
		attribute.String("block.number", blockNumberValue(blockNumber)),
	), callMsgAttributes(call)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.CallContract", attrs...)
	defer span.End()

	result, err := t.base.CallContract(ctx, call, blockNumber)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(result)),
	)

	return result, nil
}

func (t *tracingConn) CallContractAtHash(ctx context.Context, call ethereum.CallMsg, blockHash common.Hash) ([]byte, error) {
	ctx = ensureContext(ctx)
	blockHasher, ok := t.base.(bind.BlockHashContractCaller)
	if !ok {
		return nil, bind.ErrNoBlockHashState
	}

	attrs := append(t.spanAttributes("CallContractAtHash",
		attribute.String("block.hash", blockHash.Hex()),
	), callMsgAttributes(call)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.CallContractAtHash", attrs...)
	defer span.End()

	result, err := blockHasher.CallContractAtHash(ctx, call, blockHash)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(result)),
	)

	return result, nil
}

func (t *tracingConn) PendingCallContract(ctx context.Context, call ethereum.CallMsg) ([]byte, error) {
	ctx = ensureContext(ctx)
	pendingCaller, ok := t.base.(bind.PendingContractCaller)
	if !ok {
		return nil, bind.ErrNoPendingState
	}

	attrs := append(t.spanAttributes("PendingCallContract"), callMsgAttributes(call)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.PendingCallContract", attrs...)
	defer span.End()

	result, err := pendingCaller.PendingCallContract(ctx, call)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(result)),
	)

	return result, nil
}

func (t *tracingConn) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.HeaderByNumber",
		t.spanAttributes("HeaderByNumber",
			attribute.String("block.number", blockNumberValue(number)),
		)...,
	)
	defer span.End()

	header, err := t.base.HeaderByNumber(ctx, number)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	if header != nil {
		tracing.SetAttributes(span,
			attribute.String("response.hash", header.Hash().Hex()),
			attribute.String("response.number", header.Number.String()),
		)
	}

	return header, nil
}

func (t *tracingConn) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.PendingCodeAt",
		t.spanAttributes("PendingCodeAt",
			tracing.AttrAddress.String(account.Hex()),
		)...,
	)
	defer span.End()

	code, err := t.base.PendingCodeAt(ctx, account)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.length", len(code)),
	)

	return code, nil
}

func (t *tracingConn) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.PendingNonceAt",
		t.spanAttributes("PendingNonceAt",
			tracing.AttrAddress.String(account.Hex()),
		)...,
	)
	defer span.End()

	nonce, err := t.base.PendingNonceAt(ctx, account)
	if err != nil {
		tracing.RecordError(span, err)
		return 0, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.nonce", strconv.FormatUint(nonce, 10)),
	)

	return nonce, nil
}

func (t *tracingConn) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.SuggestGasPrice",
		t.spanAttributes("SuggestGasPrice")...,
	)
	defer span.End()

	price, err := t.base.SuggestGasPrice(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.gas_price", price.String()),
	)

	return price, nil
}

func (t *tracingConn) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.SuggestGasTipCap",
		t.spanAttributes("SuggestGasTipCap")...,
	)
	defer span.End()

	price, err := t.base.SuggestGasTipCap(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.gas_tip_cap", price.String()),
	)

	return price, nil
}

func (t *tracingConn) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	ctx = ensureContext(ctx)
	attrs := append(t.spanAttributes("EstimateGas"), callMsgAttributes(call)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.EstimateGas", attrs...)
	defer span.End()

	gas, err := t.base.EstimateGas(ctx, call)
	if err != nil {
		tracing.RecordError(span, err)
		return 0, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.gas", strconv.FormatUint(gas, 10)),
	)

	return gas, nil
}

func (t *tracingConn) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.SendTransaction",
		t.spanAttributes("SendTransaction",
			tracing.AttrTxHash.String(tx.Hash().Hex()),
			attribute.String("tx.nonce", strconv.FormatUint(tx.Nonce(), 10)),
			attribute.String("tx.gas", strconv.FormatUint(tx.Gas(), 10)),
		)...,
	)
	defer span.End()

	if price := tx.GasPrice(); price != nil {
		tracing.SetAttributes(span, attribute.String("tx.gas_price", price.String()))
	}
	if tip := tx.GasTipCap(); tip != nil {
		tracing.SetAttributes(span, attribute.String("tx.gas_tip_cap", tip.String()))
	}
	if fee := tx.GasFeeCap(); fee != nil {
		tracing.SetAttributes(span, attribute.String("tx.gas_fee_cap", fee.String()))
	}

	if err := t.base.SendTransaction(ctx, tx); err != nil {
		tracing.RecordError(span, err)
		return err
	}

	return nil
}

func (t *tracingConn) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	ctx = ensureContext(ctx)
	attrs := append(t.spanAttributes("FilterLogs"), filterQueryAttributes(q)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.FilterLogs", attrs...)
	defer span.End()

	logs, err := t.base.FilterLogs(ctx, q)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.log_count", len(logs)),
	)

	return logs, nil
}

func (t *tracingConn) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	ctx = ensureContext(ctx)
	attrs := append(t.spanAttributes("SubscribeFilterLogs"), filterQueryAttributes(q)...)

	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.SubscribeFilterLogs", attrs...)
	defer span.End()

	sub, err := t.base.SubscribeFilterLogs(ctx, q, ch)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return sub, nil
}

func (t *tracingConn) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx = ensureContext(ctx)
	ctx, span := tracing.StartClientSpan(ctx, "evm.rpc.TransactionReceipt",
		t.spanAttributes("TransactionReceipt",
			tracing.AttrTxHash.String(txHash.Hex()),
		)...,
	)
	defer span.End()

	receipt, err := t.base.TransactionReceipt(ctx, txHash)
	if err != nil {
		if !errors.Is(err, ethereum.NotFound) {
			tracing.RecordError(span, err)
		}
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("receipt.status", int(receipt.Status)),
		attribute.String("receipt.block_number", bigIntValue(receipt.BlockNumber)),
		attribute.String("receipt.gas_used", strconv.FormatUint(receipt.GasUsed, 10)),
	)

	return receipt, nil
}

func (t *tracingConn) spanAttributes(method string, extra ...attribute.KeyValue) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		tracing.AttrChainID.Int64(int64(t.chainID)),
		tracing.AttrMethodName.String(method),
	}

	return append(attrs, extra...)
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}

func blockNumberValue(blockNumber *big.Int) string {
	if blockNumber == nil {
		return "latest"
	}

	return blockNumber.String()
}

func bigIntValue(value *big.Int) string {
	if value == nil {
		return "0"
	}

	return value.String()
}

func callMsgAttributes(call ethereum.CallMsg) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("call.from", call.From.Hex()),
		attribute.String("call.to", optionalAddressHex(call.To)),
		attribute.String("call.value", bigIntValue(call.Value)),
		attribute.String("call.gas", strconv.FormatUint(call.Gas, 10)),
		attribute.Int("call.access_list_len", len(call.AccessList)),
		attribute.Int("call.data_size", len(call.Data)),
	}

	if call.GasPrice != nil {
		attrs = append(attrs, attribute.String("call.gas_price", call.GasPrice.String()))
	}

	if call.GasFeeCap != nil {
		attrs = append(attrs, attribute.String("call.gas_fee_cap", call.GasFeeCap.String()))
	}

	if call.GasTipCap != nil {
		attrs = append(attrs, attribute.String("call.gas_tip_cap", call.GasTipCap.String()))
	}

	return attrs
}

func filterQueryAttributes(q ethereum.FilterQuery) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.Int("filter.address_count", len(q.Addresses)),
		attribute.Int("filter.topic_count", len(q.Topics)),
	}

	if q.BlockHash != nil {
		attrs = append(attrs, attribute.String("filter.block_hash", q.BlockHash.Hex()))
	}

	if q.FromBlock != nil {
		attrs = append(attrs, attribute.String("filter.from_block", q.FromBlock.String()))
	}

	if q.ToBlock != nil {
		attrs = append(attrs, attribute.String("filter.to_block", q.ToBlock.String()))
	}

	return attrs
}

func optionalAddressHex(addr *common.Address) string {
	if addr == nil {
		return ""
	}

	return addr.Hex()
}

func (t tracingDriver) GetConfigAt(opts *bind.CallOpts, timestamp *big.Int) (gen.IValSetDriverConfig, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetConfigAt",
		tracing.AttrMethodName.String("GetConfigAt"),
		attribute.Int64("timestamp", timestamp.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	config, err := t.base.GetConfigAt(opts, timestamp)
	if err != nil {
		tracing.RecordError(span, err)
		return gen.IValSetDriverConfig{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.config", fmt.Sprintf("%+v", config)),
	)

	return config, nil
}

func (t tracingDriver) GetCurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetCurrentEpoch",
		tracing.AttrMethodName.String("GetCurrentEpoch"),
	)
	defer span.End()

	opts.Context = ctx

	epoch, err := t.base.GetCurrentEpoch(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.epoch", epoch.String()),
	)

	return epoch, nil
}

func (t tracingDriver) GetCurrentEpochDuration(opts *bind.CallOpts) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetCurrentEpochDuration",
		tracing.AttrMethodName.String("GetCurrentEpochDuration"),
	)
	defer span.End()

	opts.Context = ctx

	duration, err := t.base.GetCurrentEpochDuration(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.duration", duration.String()),
	)

	return duration, nil
}

func (t tracingDriver) GetEpochDuration(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetEpochDuration",
		tracing.AttrMethodName.String("GetEpochDuration"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	duration, err := t.base.GetEpochDuration(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.duration", duration.String()),
	)

	return duration, nil
}

func (t tracingDriver) GetEpochStart(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetEpochStart",
		tracing.AttrMethodName.String("GetEpochStart"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	start, err := t.base.GetEpochStart(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.start", start.String()),
	)

	return start, nil
}

func (t tracingDriver) SUBNETWORK(opts *bind.CallOpts) ([32]byte, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.SUBNETWORK",
		tracing.AttrMethodName.String("SUBNETWORK"),
	)
	defer span.End()

	opts.Context = ctx

	subnetwork, err := t.base.SUBNETWORK(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return [32]byte{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.subnetwork", common.Hash(subnetwork).Hex()),
	)

	return subnetwork, nil
}

func (t tracingDriver) NETWORK(opts *bind.CallOpts) (common.Address, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.NETWORK",
		tracing.AttrMethodName.String("NETWORK"),
	)
	defer span.End()

	opts.Context = ctx

	network, err := t.base.NETWORK(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return common.Address{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.network", network.Hex()),
	)

	return network, nil
}

func (t tracingDriver) RemoveSettlement(opts *bind.TransactOpts, settlement gen.IValSetDriverCrossChainAddress) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.RemoveSettlement",
		tracing.AttrMethodName.String("RemoveSettlement"),
		tracing.AttrChainID.Int64(int64(settlement.ChainId)),
		tracing.AttrAddress.String(settlement.Addr.Hex()),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.RemoveSettlement(opts, settlement)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.txHash", tx.Hash().Hex()),
	)

	return tx, nil
}

func (t tracingDriver) AddSettlement(opts *bind.TransactOpts, settlement gen.IValSetDriverCrossChainAddress) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.AddSettlement",
		tracing.AttrMethodName.String("AddSettlement"),
		tracing.AttrChainID.Int64(int64(settlement.ChainId)),
		tracing.AttrAddress.String(settlement.Addr.Hex()),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.AddSettlement(opts, settlement)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.txHash", tx.Hash().Hex()),
	)

	return tx, nil
}

type tracingSettlement struct {
	base settlementContract
}

func (t tracingSettlement) IsValSetHeaderCommittedAt(opts *bind.CallOpts, epoch *big.Int) (bool, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.IsValSetHeaderCommittedAt",
		tracing.AttrMethodName.String("IsValSetHeaderCommittedAt"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	committed, err := t.base.IsValSetHeaderCommittedAt(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.SetAttributes(span,
		attribute.Bool("response.committed", committed),
	)

	return committed, nil
}

func (t tracingSettlement) GetValSetHeaderHash(opts *bind.CallOpts) ([32]byte, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetValSetHeaderHash",
		tracing.AttrMethodName.String("GetValSetHeaderHash"),
	)
	defer span.End()

	opts.Context = ctx

	hash, err := t.base.GetValSetHeaderHash(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return [32]byte{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.hash", common.Hash(hash).Hex()),
	)

	return hash, nil
}

func (t tracingSettlement) GetValSetHeaderHashAt(opts *bind.CallOpts, epoch *big.Int) ([32]byte, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetValSetHeaderHashAt",
		tracing.AttrMethodName.String("GetValSetHeaderHashAt"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	hash, err := t.base.GetValSetHeaderHashAt(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return [32]byte{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.hash", common.Hash(hash).Hex()),
	)

	return hash, nil
}

func (t tracingSettlement) GetLastCommittedHeaderEpoch(opts *bind.CallOpts) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetLastCommittedHeaderEpoch",
		tracing.AttrMethodName.String("GetLastCommittedHeaderEpoch"),
	)
	defer span.End()

	opts.Context = ctx

	epoch, err := t.base.GetLastCommittedHeaderEpoch(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.epoch", epoch.String()),
	)

	return epoch, nil
}

func (t tracingSettlement) GetCaptureTimestampFromValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetCaptureTimestampFromValSetHeaderAt",
		tracing.AttrMethodName.String("GetCaptureTimestampFromValSetHeaderAt"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	timestamp, err := t.base.GetCaptureTimestampFromValSetHeaderAt(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.timestamp", timestamp.String()),
	)

	return timestamp, nil
}

func (t tracingSettlement) GetValSetHeaderAt(opts *bind.CallOpts, epoch *big.Int) (gen.ISettlementValSetHeader, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetValSetHeaderAt",
		tracing.AttrMethodName.String("GetValSetHeaderAt"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	header, err := t.base.GetValSetHeaderAt(opts, epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return gen.ISettlementValSetHeader{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.header", fmt.Sprintf("%+v", header)),
	)

	return header, nil
}

func (t tracingSettlement) GetValSetHeader(opts *bind.CallOpts) (gen.ISettlementValSetHeader, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetValSetHeader",
		tracing.AttrMethodName.String("GetValSetHeader"),
	)
	defer span.End()

	opts.Context = ctx

	header, err := t.base.GetValSetHeader(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return gen.ISettlementValSetHeader{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.header", fmt.Sprintf("%+v", header)),
	)

	return header, nil
}

func (t tracingSettlement) Eip712Domain(opts *bind.CallOpts) (symbiotic.Eip712Domain, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.Eip712Domain",
		tracing.AttrMethodName.String("Eip712Domain"),
	)
	defer span.End()

	opts.Context = ctx

	domain, err := t.base.Eip712Domain(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.Eip712Domain{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.domain", fmt.Sprintf("%+v", domain)),
	)

	return domain, nil
}

func (t tracingSettlement) CommitValSetHeader(opts *bind.TransactOpts, header gen.ISettlementValSetHeader, extraData []gen.ISettlementExtraData, proof []byte) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.CommitValSetHeader",
		tracing.AttrMethodName.String("CommitValSetHeader"),
		tracing.AttrEpoch.Int64(header.Epoch.Int64()),
		tracing.AttrProofSize.Int(len(proof)),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.CommitValSetHeader(opts, header, extraData, proof)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

func (t tracingSettlement) SetGenesis(opts *bind.TransactOpts, valSetHeader gen.ISettlementValSetHeader, extraData []gen.ISettlementExtraData) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.SetGenesis",
		tracing.AttrMethodName.String("SetGenesis"),
		tracing.AttrEpoch.Int64(valSetHeader.Epoch.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.SetGenesis(opts, valSetHeader, extraData)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

func (t tracingSettlement) VerifyQuorumSigAt(opts *bind.CallOpts, message []byte, keyTag uint8, quorumThreshold *big.Int, proof []byte, epoch *big.Int, hint []byte) (bool, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.VerifyQuorumSigAt",
		tracing.AttrMethodName.String("VerifyQuorumSigAt"),
		tracing.AttrEpoch.Int64(epoch.Int64()),
		tracing.AttrKeyTag.String(strconv.FormatUint(uint64(keyTag), 10)),
		tracing.AttrProofSize.Int(len(proof)),
	)
	defer span.End()

	opts.Context = ctx

	valid, err := t.base.VerifyQuorumSigAt(opts, message, keyTag, quorumThreshold, proof, epoch, hint)
	if err != nil {
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.SetAttributes(span,
		attribute.Bool("response.valid", valid),
	)

	return valid, nil
}

type tracingVotingPowerProvider struct {
	base votingPowerProviderContract
}

func (t tracingVotingPowerProvider) Nonces(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.Nonces",
		tracing.AttrMethodName.String("Nonces"),
		tracing.AttrAddress.String(owner.Hex()),
	)
	defer span.End()

	opts.Context = ctx

	nonce, err := t.base.Nonces(opts, owner)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.nonce", nonce.String()),
	)

	return nonce, nil
}

func (t tracingVotingPowerProvider) GetVotingPowersAt(opts *bind.CallOpts, extraData [][]byte, timestamp *big.Int) ([]gen.IVotingPowerProviderOperatorVotingPower, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetVotingPowersAt",
		tracing.AttrMethodName.String("GetVotingPowersAt"),
		attribute.Int64("timestamp", timestamp.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	powers, err := t.base.GetVotingPowersAt(opts, extraData, timestamp)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.operators_count", len(powers)),
	)

	return powers, nil
}

func (t tracingVotingPowerProvider) GetOperatorsAt(opts *bind.CallOpts, timestamp *big.Int) ([]common.Address, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetOperatorsAt",
		tracing.AttrMethodName.String("GetOperatorsAt"),
		attribute.Int64("timestamp", timestamp.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	operators, err := t.base.GetOperatorsAt(opts, timestamp)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.operators_count", len(operators)),
	)

	return operators, nil
}

func (t tracingVotingPowerProvider) Eip712Domain(opts *bind.CallOpts) (symbiotic.Eip712Domain, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.VotingPowerProvider_Eip712Domain",
		tracing.AttrMethodName.String("Eip712Domain"),
	)
	defer span.End()

	opts.Context = ctx

	domain, err := t.base.Eip712Domain(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.Eip712Domain{}, err
	}

	tracing.SetAttributes(span,
		attribute.String("response.domain", fmt.Sprintf("%+v", domain)),
	)

	return domain, nil
}

func (t tracingVotingPowerProvider) IsOperatorRegistered(opts *bind.CallOpts, operator common.Address) (bool, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.IsOperatorRegistered",
		tracing.AttrMethodName.String("IsOperatorRegistered"),
		tracing.AttrAddress.String(operator.Hex()),
	)
	defer span.End()

	opts.Context = ctx

	registered, err := t.base.IsOperatorRegistered(opts, operator)
	if err != nil {
		tracing.RecordError(span, err)
		return false, err
	}

	tracing.SetAttributes(span,
		attribute.Bool("response.registered", registered),
	)

	return registered, nil
}

type tracingVotingPowerProviderTransactor struct {
	base votingPowerProviderTransactor
}

func (t tracingVotingPowerProviderTransactor) InvalidateOldSignatures(opts *bind.TransactOpts) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.InvalidateOldSignatures",
		tracing.AttrMethodName.String("InvalidateOldSignatures"),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.InvalidateOldSignatures(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

func (t tracingVotingPowerProviderTransactor) RegisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.VotingPowerProvider_RegisterOperator",
		tracing.AttrMethodName.String("RegisterOperator"),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.RegisterOperator(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

func (t tracingVotingPowerProviderTransactor) UnregisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.UnregisterOperator",
		tracing.AttrMethodName.String("UnregisterOperator"),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.UnregisterOperator(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

type tracingKeyRegistry struct {
	base keyRegistryContract
}

func (t tracingKeyRegistry) GetKeysOperatorsAt(opts *bind.CallOpts, timestamp *big.Int) ([]common.Address, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetKeysOperatorsAt",
		tracing.AttrMethodName.String("GetKeysOperatorsAt"),
		attribute.Int64("timestamp", timestamp.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	operators, err := t.base.GetKeysOperatorsAt(opts, timestamp)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.operators_count", len(operators)),
	)

	return operators, nil
}

func (t tracingKeyRegistry) GetKeysAt(opts *bind.CallOpts, timestamp *big.Int) ([]gen.IKeyRegistryOperatorWithKeys, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.GetKeysAt",
		tracing.AttrMethodName.String("GetKeysAt"),
		attribute.Int64("timestamp", timestamp.Int64()),
	)
	defer span.End()

	opts.Context = ctx

	keys, err := t.base.GetKeysAt(opts, timestamp)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		attribute.Int("response.operators_count", len(keys)),
	)

	return keys, nil
}

func (t tracingKeyRegistry) SetKey(opts *bind.TransactOpts, tag uint8, key []byte, signature []byte, extraData []byte) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.SetKey",
		tracing.AttrMethodName.String("SetKey"),
		tracing.AttrKeyTag.String(strconv.FormatUint(uint64(tag), 10)),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.SetKey(opts, tag, key, signature, extraData)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}

type tracingOperatorRegistry struct {
	base operatorRegistryContract
}

func (t tracingOperatorRegistry) RegisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := tracing.StartClientSpan(ctx, "evm.OperatorRegistry_RegisterOperator",
		tracing.AttrMethodName.String("RegisterOperator"),
	)
	defer span.End()

	opts.Context = ctx

	tx, err := t.base.RegisterOperator(opts)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.AttrTxHash.String(tx.Hash().Hex()),
	)

	return tx, nil
}
