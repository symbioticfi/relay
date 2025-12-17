package evm

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/e2e/tests/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type Config struct {
	ChainURL   string
	PrivateKey symbiotic.PrivateKey
}

type Client struct {
	client  *ethclient.Client
	cfg     Config
	chainID *big.Int
}

func NewClient(t *testing.T, cfg Config) *Client {
	t.Helper()
	client, err := ethclient.DialContext(t.Context(), cfg.ChainURL)
	require.NoError(t, err)
	chainID, err := client.ChainID(t.Context())
	require.NoError(t, err)

	return &Client{
		client:  client,
		chainID: chainID,
		cfg:     cfg,
	}
}

func (e *Client) TransferMockToken(ctx context.Context, tokenAddress, to common.Address, amount *big.Int) (symbiotic.TxResult, error) {
	erc20, err := gen.NewMockERC20(tokenAddress, e.client)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to instantiate mock erc20: %w", err)
	}

	return e.doTransaction(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return erc20.Transfer(opts, to, amount)
	})
}

func (e *Client) ApproveMockToken(ctx context.Context, tokenAddress, vault common.Address, amount *big.Int) (symbiotic.TxResult, error) {
	erc20, err := gen.NewMockERC20(tokenAddress, e.client)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to instantiate mock erc20: %w", err)
	}

	return e.doTransaction(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return erc20.Approve(opts, vault, amount)
	})
}

func (e *Client) VaultDeposit(ctx context.Context, vaultAddress, tokenAddress common.Address, amount *big.Int) (symbiotic.TxResult, error) {
	vault, err := gen.NewVault(vaultAddress, e.client)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to instantiate mock erc20: %w", err)
	}

	return e.doTransaction(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return vault.Deposit(opts, tokenAddress, amount)
	})
}

func (e *Client) OptIn(ctx context.Context, optInServiceAddr, networkAddress common.Address) (symbiotic.TxResult, error) {
	optInService, err := gen.NewOptInService(optInServiceAddr, e.client)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to instantiate mock optInService: %w", err)
	}

	return e.doTransaction(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return optInService.OptIn(opts, networkAddress)
	})
}

func (e *Client) GetAutoDeployVault(ctx context.Context, autoDeployVaultAddress, operator common.Address) (common.Address, error) {
	autoDeploy, err := gen.NewOpNetVaultAutoDeployLogic(autoDeployVaultAddress, e.client)
	if err != nil {
		return common.Address{}, errors.Errorf("failed to instantiate auto deploy: %w", err)
	}

	vaultAddress, err := autoDeploy.GetAutoDeployedVault(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     ctx,
	}, operator)
	if err != nil {
		return common.Address{}, errors.Errorf("failed to get auto deployed vault: %w", err)
	}

	return vaultAddress, nil
}

func (e *Client) IsAutoDeployEnabled(ctx context.Context, autoDeployVaultAddress common.Address) (bool, error) {
	autoDeploy, err := gen.NewOpNetVaultAutoDeployLogic(autoDeployVaultAddress, e.client)
	if err != nil {
		return false, errors.Errorf("failed to instantiate auto deploy: %w", err)
	}

	enabled, err := autoDeploy.IsAutoDeployEnabled(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     ctx,
	})
	if err != nil {
		return false, errors.Errorf("failed to get auto deployed vault: %w", err)
	}

	return enabled, nil
}

func (e *Client) doTransaction(ctx context.Context, f func(opts *bind.TransactOpts) (*types.Transaction, error)) (symbiotic.TxResult, error) {
	ecdsaKey, err := crypto.ToECDSA(e.cfg.PrivateKey.Bytes())
	if err != nil {
		return symbiotic.TxResult{}, err
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(ecdsaKey, e.chainID)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	tmCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	txOpts.Context = tmCtx

	tx, err := f(txOpts)
	if err != nil {
		return symbiotic.TxResult{}, err
	}

	receipt, err := bind.WaitMined(ctx, e.client, tx)
	if err != nil {
		return symbiotic.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return symbiotic.TxResult{}, errors.New("transaction reverted on chain")
	}

	return symbiotic.TxResult{
		TxHash:            receipt.TxHash,
		GasUsed:           receipt.GasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
	}, nil
}
