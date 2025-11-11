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

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) CommitValsetHeader(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
	header symbiotic.ValidatorSetHeader,
	extraData []symbiotic.ExtraData,
	proof []byte,
) (_ symbiotic.TxResult, err error) {
	pk, err := e.cfg.KeyProvider.GetPrivateKeyByNamespaceTypeId(
		keyprovider.EVM_KEY_NAMESPACE,
		symbiotic.KeyTypeEcdsaSecp256k1,
		int(addr.ChainId),
	)
	if err != nil {
		return symbiotic.TxResult{}, err
	}
	ecdsaKey, err := crypto.ToECDSA(pk.Bytes())
	if err != nil {
		return symbiotic.TxResult{}, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(ecdsaKey, new(big.Int).SetUint64(addr.ChainId))
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("CommitValSetHeader", addr.ChainId, err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(uint64(header.Epoch)),
		CaptureTimestamp:   new(big.Int).SetUint64(uint64(header.CaptureTimestamp)),
		QuorumThreshold:    header.QuorumThreshold.Int,
		TotalVotingPower:   header.TotalVotingPower.Int,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
	}

	extraDataDTO := make([]gen.ISettlementExtraData, len(extraData))
	for i, extraData := range extraData {
		extraDataDTO[i].Key = extraData.Key
		extraDataDTO[i].Value = extraData.Value
	}

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := settlement.CommitValSetHeader(txOpts, headerDTO, extraDataDTO, proof)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMContractError(gen.ISettlementMetaData, err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return symbiotic.TxResult{}, errors.New("transaction reverted on chain")
	}

	slog.DebugContext(ctx, "Valset header committed", "receipt", receipt)

	e.metrics.ObserveCommitValsetHeaderParams(addr.ChainId, receipt.GasUsed, receipt.EffectiveGasPrice)

	return symbiotic.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}
