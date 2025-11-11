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
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) RegisterOperator(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
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
		e.observeMetrics("RegisterOperator", addr.ChainId, err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	registry, err := e.getOperatorRegistryContract(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := registry.RegisterOperator(txOpts)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMError(err)
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

func (e *Client) RegisterKey(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
	keyTag symbiotic.KeyTag,
	key symbiotic.CompactPublicKey,
	signature symbiotic.RawSignature,
	extraData []byte,
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
		e.observeMetrics("SetKey", addr.ChainId, err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	registry, err := e.getKeyRegistryContract(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get settlement contract: %w", err)
	}

	tx, err := registry.SetKey(txOpts, uint8(keyTag), key, signature, extraData)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMError(err)
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

func (e *Client) InvalidateOldSignatures(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
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
		e.observeMetrics("InvalidateOldSignatures", addr.ChainId, err, now)
	}(time.Now())
	txOpts.Context = tmCtx

	votingPowerProvider, err := e.getVotingPowerProviderContractTransactor(addr)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to get voting power provider contract: %w", err)
	}

	tx, err := votingPowerProvider.InvalidateOldSignatures(txOpts)
	if err != nil {
		return symbiotic.TxResult{}, e.formatEVMError(err)
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
