package evm

import (
	"context"
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

func (e *Client) doTransaction(ctx context.Context, method string, addr symbiotic.CrossChainAddress, f func(opts *bind.TransactOpts) (*types.Transaction, error)) (symbiotic.TxResult, error) {
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

	tx, err := f(txOpts)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMContractError(gen.SettlementMetaData, err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return symbiotic.TxResult{}, errors.New("transaction reverted on chain")
	}

	return symbiotic.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}
