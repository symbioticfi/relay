package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"

	"middleware-offchain/core/client/evm/gen"
	"middleware-offchain/core/entity"
)

func (e *Client) CommitValsetHeader(ctx context.Context, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, new(big.Int).SetUint64(e.masterAddress.ChainId))
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(header.Epoch),
		CaptureTimestamp:   new(big.Int).SetUint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}

	extraDataDTO := make([]gen.ISettlementExtraData, len(extraData))
	for i, extraData := range extraData {
		extraDataDTO[i].Key = extraData.Key
		extraDataDTO[i].Value = extraData.Value
	}

	tx, err := e.settlement.CommitValSetHeader(txOpts, headerDTO, extraDataDTO, proof)
	if err != nil {
		return entity.TxResult{}, e.formatEVMError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.client, tx)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return entity.TxResult{}, errors.New("transaction reverted on chain")
	}

	return entity.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}
