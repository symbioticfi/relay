package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rpc"

	"middleware-offchain/core/entity"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, epoch uint64, message []byte, keyTag entity.KeyTag, threshold *big.Int, proof []byte) (bool, error) {
	toCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	result, err := e.settlement.VerifyQuorumSigAt(&bind.CallOpts{
		BlockNumber: new(big.Int).SetInt64(rpc.FinalizedBlockNumber.Int64()),
		Context:     toCtx,
	}, message, uint8(keyTag), threshold, proof, new(big.Int).SetUint64(epoch), []byte{})

	if err != nil {
		return false, err
	}

	return result, nil
}
