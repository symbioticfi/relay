package symbiotic

import (
	"context"
	"fmt"
	"math/big"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, epoch *big.Int, message []byte, keyTag uint8, threshold *big.Int, proof, hint []byte) (bool, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, verifyQuorumSigFunction, epoch, message, keyTag, threshold, proof, hint)
	if err != nil {
		return false, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, fmt.Errorf("failed to call contract: %w", err)
	}

	return new(big.Int).SetBytes(result).Cmp(big.NewInt(1)) == 0, nil
}
