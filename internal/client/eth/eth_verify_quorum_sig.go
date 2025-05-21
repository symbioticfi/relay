package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, message []byte, keyTag uint8, threshold *big.Int, proof []byte) (bool, error) {
	callMsg, err := constructCallMsg(e.masterContractAddress, masterABI, verifyQuorumSigFunction, message, keyTag, threshold, proof)
	if err != nil {
		return false, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, callMsg)
	if err != nil {
		return false, fmt.Errorf("failed to call contract: %w", err)
	}

	return *abi.ConvertType(result[0], new(bool)).(*bool), nil
}
