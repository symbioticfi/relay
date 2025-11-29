package evm

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/pkg/tracing"
	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type tracingDriver struct {
	base driverContract
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
