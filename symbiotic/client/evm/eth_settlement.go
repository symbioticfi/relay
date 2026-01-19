package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) RemoveSettlement(
	ctx context.Context,
	settlementAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	return e.doTransaction(ctx, "RemoveSettlement", e.cfg.DriverAddress, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return e.driver.RemoveSettlement(txOpts, gen.IValSetDriverCrossChainAddress{
			ChainId: settlementAddr.ChainId,
			Addr:    settlementAddr.Address,
		})
	})
}

func (e *Client) AddSettlement(
	ctx context.Context,
	settlementAddr symbiotic.CrossChainAddress,
) (_ symbiotic.TxResult, err error) {
	return e.doTransaction(ctx, "AddSettlement", e.cfg.DriverAddress, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return e.driver.AddSettlement(txOpts, gen.IValSetDriverCrossChainAddress{
			ChainId: settlementAddr.ChainId,
			Addr:    settlementAddr.Address,
		})
	})
}

func (e *Client) doTransaction(ctx context.Context, method string, addr symbiotic.CrossChainAddress, f func(opts *bind.TransactOpts) (*types.Transaction, error), opts ...symbiotic.EVMOption) (symbiotic.TxResult, error) {
	evmOpts := symbiotic.AppliedEVMOptions(opts...)

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
		e.observeMetrics(method, e.driverChainID, err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	if !e.conns[addr.ChainId].hasMaxPriorityFeePerGasMethod {
		txOpts.GasPrice = big.NewInt(2_000_000_000) // 2 GWei
	}

	// If GasLimitMultiplier is set, estimate gas and apply multiplier
	if evmOpts.GasLimitMultiplier > 0 {
		txOpts.NoSend = true
		dryRunTx, err := f(txOpts)
		if err != nil {
			return symbiotic.TxResult{}, e.formatEVMError(err)
		}

		msg := ethereum.CallMsg{
			From:  txOpts.From,
			To:    dryRunTx.To(),
			Data:  dryRunTx.Data(),
			Value: dryRunTx.Value(),
		}
		estimatedGas, err := e.conns[addr.ChainId].EstimateGas(tmCtx, msg)
		if err != nil {
			return symbiotic.TxResult{}, errors.Errorf("failed to estimate gas: %w", err)
		}

		txOpts.GasLimit = uint64(float64(estimatedGas) * evmOpts.GasLimitMultiplier)
		txOpts.NoSend = false
	}

	tx, err := f(txOpts)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return symbiotic.TxResult{
			TxHash:            receipt.TxHash,
			GasUsed:           receipt.GasUsed,
			EffectiveGasPrice: receipt.EffectiveGasPrice,
		}, errors.Errorf("transaction %s reverted on chain (gasUsed: %d)", receipt.TxHash.Hex(), receipt.GasUsed)
	}

	return symbiotic.TxResult{
		TxHash:            receipt.TxHash,
		GasUsed:           receipt.GasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
	}, nil
}
