package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract related constants
const (
	contractABI = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"messageHash","type":"bytes32"},{"indexed":false,"internalType":"bytes","name":"signature","type":"bytes"},{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"AggregatedSignatureSubmitted","type":"event"},{"inputs":[{"internalType":"bytes32","name":"messageHash","type":"bytes32"},{"internalType":"bytes","name":"signature","type":"bytes"}],"name":"submitAggregatedSignature","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"processedMessages","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
)

type Key struct {
	Tag     uint8
	Payload []byte
}

type Vault struct {
	VaultAddress common.Address
	VotingPower  *big.Int
}

type Validator struct {
	Operator    common.Address
	VotingPower *big.Int
	IsActive    bool
	Keys        []Key
	Vaults      []Vault
}

type ValidatorSet struct {
	TotalActiveVotingPower *big.Int
	Validators             []Validator
}

type ETHService struct {
	ethClient       *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	storage         *Storage
	epochStart      *big.Int
	epochDuration   *big.Int
	lastEpoch       *big.Int
}

func NewETHService(rpcUrl string, contractAddress string, lastEpoch *big.Int, storage *Storage) (*ETHService, error) {
	contractABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// get epoch start and duration

	return &ETHService{
		ethClient:       ethClient,
		contractAddress: common.HexToAddress(contractAddress),
		contractABI:     contractABI,
		lastEpoch:       lastEpoch,
		storage:         storage,
	}, nil
}

func (e *ETHService) Commit(messageHash string, signature []byte) error {
	return nil
}

func (e *ETHService) getValidatorSet(epoch *big.Int) (ValidatorSet, error) {
	// This is a mock implementation of getValidatorSet
	// In a real implementation, this would query the Ethereum contract
	// to get the current validator set for the given epoch

	log.Printf("Getting validator set for epoch %s", epoch.String())

	// Create a mock validator set with some test data
	mockValidatorSet := ValidatorSet{
		TotalActiveVotingPower: big.NewInt(1000),
		Validators: []Validator{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: big.NewInt(400),
				IsActive:    true,
				Keys: []Key{
					{
						Tag:     1,
						Payload: []byte("validator1pubkey"),
					},
				},
				Vaults: []Vault{
					{
						VaultAddress: common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
						VotingPower:  big.NewInt(400),
					},
				},
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: big.NewInt(300),
				IsActive:    true,
				Keys: []Key{
					{
						Tag:     2,
						Payload: []byte("validator2pubkey"),
					},
				},
				Vaults: []Vault{
					{
						VaultAddress: common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
						VotingPower:  big.NewInt(300),
					},
				},
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: big.NewInt(300),
				IsActive:    true,
				Keys: []Key{
					{
						Tag:     3,
						Payload: []byte("validator3pubkey"),
					},
				},
				Vaults: []Vault{
					{
						VaultAddress: common.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc"),
						VotingPower:  big.NewInt(300),
					},
				},
			},
			{
				Operator:    common.HexToAddress("0x4444444444444444444444444444444444444444"),
				VotingPower: big.NewInt(0),
				IsActive:    false,
				Keys: []Key{
					{
						Tag:     4,
						Payload: []byte("validator4pubkey"),
					},
				},
				Vaults: []Vault{},
			},
		},
	}

	return mockValidatorSet, nil

}

func (e *ETHService) GetCurrentEpoch() (*big.Int, error) {
	return e.lastEpoch, nil
}

func (e *ETHService) checkAndSubmitSignatures(ctx context.Context) error {
	return nil
}

func (e *ETHService) checkNewValidatorSet(ctx context.Context) error {
	return nil
}

// CheckSignatures checks for signatures in storage for the current epoch at regular intervals
func (e *ETHService) Start(ctx context.Context, storage *Storage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.checkNewValidatorSet(ctx)
			e.checkAndSubmitSignatures(ctx)
		}
	}
}
