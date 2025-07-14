package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
	keyprovider "middleware-offchain/core/usecase/key-provider"
)

func (e *Client) RegisterOperator(
	ctx context.Context,
	addr entity.CrossChainAddress,
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
		e.observeMetrics("RegisterOperator", err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	registry, err := e.getOperatorRegistryContract(addr)
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
		e.observeMetrics("SetKey", err, now)
	}(time.Now())
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
