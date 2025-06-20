package evm

import (
	"context"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"

	"middleware-offchain/core/client/evm/gen"
	"middleware-offchain/core/entity"
)

func (e *Client) CommitValsetHeader(
	ctx context.Context,
	addr entity.CrossChainAddress,
	header entity.ValidatorSetHeader,
	extraData []entity.ExtraData,
	proof []byte,
) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, new(big.Int).SetUint64(e.cfg.DriverAddress.ChainId))
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(header.Epoch),
		CaptureTimestamp:   new(big.Int).SetUint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold.Int,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
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

	return entity.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}
