package symbiotic

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-errors/errors"
	"golang.org/x/crypto/sha3"

	"middleware-offchain/core/client/symbiotic/gen"
	"middleware-offchain/core/entity"
)

func (e *Client) CommitValsetHeader(ctx context.Context, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, big.NewInt(111)) // todo ilya, pass chain id from config
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(header.Epoch),
		CaptureTimestamp:   new(big.Int).SetUint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}

	extraDataDTO := make([]gen.ISettlementExtraData, len(extraData))
	for i, extraData := range extraData {
		extraDataDTO[i].Key = extraData.Key
		extraDataDTO[i].Value = extraData.Value
	}

	tx, err := e.master.CommitValSetHeader(txOpts, headerDTO, extraDataDTO, proof, []byte{})
	if err != nil {
		return entity.TxResult{}, e.formatEthError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.client, tx)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return entity.TxResult{}, errors.New("transaction reverted on chain")
	}

	return entity.TxResult{
		TxHash: receipt.TxHash,
	}, nil
}

func (e *Client) SetGenesis(ctx context.Context, header entity.ValidatorSetHeader, extraData []entity.ExtraData) (entity.TxResult, error) {
	if e.masterPK == nil {
		return entity.TxResult{}, errors.New("master private key is not set")
	}
	tmCtx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	txOpts, err := bind.NewKeyedTransactorWithChainID(e.masterPK, big.NewInt(111)) // todo ilya, pass chain id from config
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to create new keyed transactor: %w", err)
	}
	txOpts.Context = tmCtx

	headerDTO := gen.ISettlementValSetHeader{
		Version:            header.Version,
		RequiredKeyTag:     uint8(header.RequiredKeyTag),
		Epoch:              new(big.Int).SetUint64(header.Epoch),
		CaptureTimestamp:   new(big.Int).SetUint64(header.CaptureTimestamp),
		QuorumThreshold:    header.QuorumThreshold,
		ValidatorsSszMRoot: header.ValidatorsSszMRoot,
		PreviousHeaderHash: header.PreviousHeaderHash,
	}

	extraDataDTO := make([]gen.ISettlementExtraData, len(extraData))
	for i, extraData := range extraData {
		extraDataDTO[i].Key = extraData.Key
		extraDataDTO[i].Value = extraData.Value
	}

	tx, err := e.master.SetGenesis(txOpts, headerDTO, extraDataDTO)
	if err != nil {
		return entity.TxResult{}, e.formatEthError(err)
	}

	receipt, err := bind.WaitMined(ctx, e.client, tx)
	if err != nil {
		return entity.TxResult{}, errors.Errorf("failed to wait for tx mining: %w", err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return entity.TxResult{}, errors.New("transaction reverted on chain")
	}

	return entity.TxResult{
		TxHash: receipt.TxHash,
	}, nil
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
