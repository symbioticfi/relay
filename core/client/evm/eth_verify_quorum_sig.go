package evm

import (
	"context"
	"math/big"

	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, epoch uint64, message []byte, keyTag entity.KeyTag, threshold *big.Int, proof []byte) (bool, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, verifyQuorumSigFunction, new(big.Int).SetUint64(epoch), message, uint8(keyTag), threshold, proof, []byte{})
	if err != nil {
		return false, errors.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, errors.Errorf("failed to call contract: %w", err)
	}

	return new(big.Int).SetBytes(result).Cmp(big.NewInt(1)) == 0, nil
}
