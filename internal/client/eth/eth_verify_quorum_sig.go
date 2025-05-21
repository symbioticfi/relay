package eth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"golang.org/x/crypto/sha3"
)

func (e *Client) VerifyQuorumSig(ctx context.Context, message []byte, keyTag uint8, threshold *big.Int, proof []byte) error {
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

	tx, err := e.master.VerifyQuorumSig(txOpts, message, keyTag, threshold, proof)
	if err != nil {
		return e.formatEthError(err)
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

func (e *Client) formatEthError(err error) error {
	type jsonError interface {
		ErrorData() interface{}
		ErrorCode() int
	}
	var errData jsonError
	if !errors.As(err, &errData) {
		return errors.Errorf("failed to commit valset header: %w", err)
	}
	if errData.ErrorCode() != 3 && errData.ErrorData() == nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	errSelector, ok := errData.ErrorData().(string)
	if !ok {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	for name, errDef := range masterABI.Errors {
		selector := keccak4(errDef.String())
		if "0x"+selector == errSelector {
			// SettlementManager_VerificationFailed();
			// 0xfb6101e9
			fmt.Printf("🧩 Найдено совпадение: %s => 0x%s\n", errDef.Sig, selector) //nolint:forbidigo // todo ilya fix
			return errors.Errorf("failed to commit valset header: %s", name)
		}
	}

	masterErr, ok := masterABI.Errors[errSelector]
	if !ok {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	return errors.Errorf("failed to commit valset header: %s", masterErr)
}

// keccak256(signature)[:4]
func keccak4(sig string) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(sig))
	return hex.EncodeToString(hash.Sum(nil)[:4])
}
