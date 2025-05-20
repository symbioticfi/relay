package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"golang.org/x/crypto/sha3"

	"middleware-offchain/internal/client/eth/gen"
	"middleware-offchain/internal/entity"
)

func (e *Client) CommitValsetHeader2(ctx context.Context, header entity.ValidatorSetHeader, proof []byte) error {
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

	pack, err := masterABI.Pack("commitValSetHeader", headerDTO, proof)
	if err != nil {
		return errors.Errorf("failed to pack commit valset header: %w", err)
	}

	pk, err := crypto.ToECDSA(e.cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key: %w", err)
	}

	// Получение адреса отправителя
	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.Errorf("Failed to cast public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Получение nonce
	nonce, err := e.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return errors.Errorf("Failed to get nonce: %v", err)
	}

	// Установка параметров транзакции
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return errors.Errorf("Failed to get gas price: %v", err)
	}

	chainID, err := e.client.NetworkID(ctx)
	if err != nil {
		return errors.Errorf("Failed to get chain ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	if err != nil {
		return errors.Errorf("Failed to create transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // Сумма в wei (0 для вызова метода контракта)
	auth.GasLimit = uint64(300000) // Лимит газа
	auth.GasPrice = gasPrice

	tx := types.NewTransaction(nonce, e.masterContractAddress, big.NewInt(0), uint64(300000), gasPrice, pack)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), pk)
	if err != nil {
		return errors.Errorf("Failed to sign transaction: %v", err)
	}

	// Отправка транзакции
	err = e.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return errors.Errorf("Failed to send transaction: %v", err)
	}

	return nil
}

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

	pack, err := masterABI.Pack("commitValSetHeader", headerDTO, proof)
	if err != nil {
		return errors.Errorf("failed to pack commit valset header: %w", err)
	}

	fmt.Println("header.Version>>>", header.Version)
	fmt.Println("header.ActiveAggregatedKeys[0].Tag>>>", header.ActiveAggregatedKeys[0].Tag)
	fmt.Println("hex.EncodeToString(header.ActiveAggregatedKeys[0].Payload)>>>", hex.EncodeToString(header.ActiveAggregatedKeys[0].Payload))
	fmt.Println("header.TotalActiveVotingPower", header.TotalActiveVotingPower.String())
	fmt.Println("header.ValidatorsSszMRoot", hex.EncodeToString(header.ValidatorsSszMRoot[:]))
	fmt.Println("header.ExtraData", hex.EncodeToString(header.ExtraData))
	fmt.Println("proof", hex.EncodeToString(proof))
	fmt.Println("pack", hex.EncodeToString(pack))

	tx, err := e.master.CommitValSetHeader(txOpts, headerDTO, proof)
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
