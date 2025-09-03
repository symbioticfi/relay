package evm

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/client/evm/gen"
	"github.com/symbioticfi/relay/core/entity"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
)

func (e *Client) CommitValsetHeader(
	ctx context.Context,
	addr entity.CrossChainAddress,
	header entity.ValidatorSetHeader,
	extraData []entity.ExtraData,
	proof []byte,
) (_ entity.TxResult, err error) {
	pk, err := e.cfg.KeyProvider.GetPrivateKeyByNamespaceTypeId(
		keyprovider.EVM_KEY_NAMESPACE,
		entity.KeyTypeEcdsaSecp256k1,
		int(addr.ChainId),
	)
	if err != nil {
		return entity.TxResult{}, err
	}
	ecdsaKey, err := crypto.ToECDSA(pk.Bytes())
	if err != nil {
		return entity.TxResult{}, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(ecdsaKey, new(big.Int).SetUint64(addr.ChainId))
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("CommitValSetHeader", err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(header.Epoch),
		CaptureTimestamp:   new(big.Int).SetUint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold.Int,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
	}

	extraDataDTO := make([]gen.ISettlementExtraData, len(extraData))
	for i, extraData := range extraData {
		extraDataDTO[i].Key = extraData.Key
		extraDataDTO[i].Value = extraData.Value
	}

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := settlement.CommitValSetHeader(txOpts, headerDTO, extraDataDTO, proof)
	if err != nil {
		return entity.TxResult{}, e.formatEVMContractError(gen.ISettlementMetaData, err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return entity.TxResult{}, errors.New("transaction reverted on chain")
	}

	slog.DebugContext(ctx, "Valset header committed", "receipt", receipt)

	e.metrics.ObserveCommitValsetHeaderParams(addr.ChainId, receipt.GasUsed, receipt.EffectiveGasPrice)

	return entity.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}
