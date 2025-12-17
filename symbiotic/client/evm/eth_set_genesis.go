package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/symbiotic/client/evm/gen"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) SetGenesis(
	ctx context.Context,
	addr symbiotic.CrossChainAddress,
	header symbiotic.ValidatorSetHeader,
	extraData []symbiotic.ExtraData,
) (_ symbiotic.TxResult, err error) {
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

	return e.doTransaction(ctx, "SetGenesis", addr.ChainId, func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
		return settlement.SetGenesis(txOpts, headerDTO, extraDataDTO)
	})
}
