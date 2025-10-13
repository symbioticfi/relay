package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, addr symbiotic.CrossChainAddress, epoch symbiotic.Epoch, message []byte, keyTag symbiotic.KeyTag, threshold *big.Int, proof []byte) (_ bool, err error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()
	defer func(now time.Time) {
		e.observeMetrics("VerifyQuorumSigAt", err, now)
	}(time.Now())

	settlement, err := e.getSettlementContract(addr)
	if err != nil {
		return false, errors.Errorf("failed to get settlement contract: %w", err)
	}

	result, err := settlement.VerifyQuorumSigAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, message, uint8(keyTag), threshold, proof, new(big.Int).SetUint64(uint64(epoch)), []byte{})

	if err != nil {
		return false, err
	}

	return result, nil
}
