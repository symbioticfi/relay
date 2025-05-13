package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/internal/client/eth/gen"
	"middleware-offchain/internal/entity"
)

func (e *Client) CommitValsetHeader(ctx context.Context, header entity.ValidatorSetHeader, proof []byte) error {
	if e.masterPK == nil {
		return errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, big.NewInt(111)) // todo ilya, pass chain id from config
	if err != nil {
		return errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementManagerValSetHeader{
		Version: header.Version,
		ActiveAggregatedKeys: lo.Map(header.ActiveAggregatedKeys, func(key entity.Key, _ int) gen.IBaseKeyManagerKey {
			return gen.IBaseKeyManagerKey{
				Tag:     key.Tag,
				Payload: key.Payload,
			}
		}),
		TotalActiveVotingPower: header.TotalActiveVotingPower,
		ValidatorsSszMRoot:     header.ValidatorsSszMRoot,
		ExtraData:              header.ExtraData,
	}

	tx, err := e.master.CommitValSetHeader(txOpts, headerDTO, proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, e.client, tx)
	if err != nil {
		return errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return errors.New("transaction reverted on chain")
	}

	return nil
}
