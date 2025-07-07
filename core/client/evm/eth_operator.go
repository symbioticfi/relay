package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

func (e *Client) RegisterOperator(
	ctx context.Context,
	addr entity.CrossChainAddress,
) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, new(big.Int).SetUint64(addr.ChainId))
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	registry, err := e.getOperagorRegistryContract(addr)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := registry.RegisterOperator(txOpts)
	if err != nil {
		return entity.TxResult{}, e.formatEVMError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
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

func (e *Client) RegisterKey(
	ctx context.Context,
	addr entity.CrossChainAddress,
	keyTag entity.KeyTag,
	key entity.CompactPublicKey,
	signature entity.RawSignature,
	extraData []byte,
) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, new(big.Int).SetUint64(addr.ChainId))
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	registry, err := e.getKeyRegistryContract(addr)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := registry.SetKey(txOpts, uint8(keyTag), key, signature, extraData)
	if err != nil {
		return entity.TxResult{}, e.formatEVMError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.conns[addr.ChainId], tx)
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
