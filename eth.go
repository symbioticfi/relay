package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract related constants
const (
	contractABI = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"messageHash","type":"bytes32"},{"indexed":false,"internalType":"bytes","name":"signature","type":"bytes"},{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"AggregatedSignatureSubmitted","type":"event"},{"inputs":[{"internalType":"bytes32","name":"messageHash","type":"bytes32"},{"internalType":"bytes","name":"signature","type":"bytes"}],"name":"submitAggregatedSignature","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"processedMessages","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
)

// Phase represents the different phases of the protocol
type Phase uint64

const (
	IDLE Phase = iota
	COMMIT
	ACCEPT
	FAIL
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
	client          *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	storage         *Storage
}

func NewETHService(rpcUrl string, contractAddress string, storage *Storage) (*ETHService, error) {
	contractABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// get epoch start and duration

	return &ETHService{
		client:          client,
		contractAddress: common.HexToAddress(contractAddress),
		contractABI:     contractABI,
		storage:         storage,
	}, nil
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

func (e *ETHService) Commit(messageHash string, signature []byte) error {
	return nil
}

func (e *ETHService) getMockValidatorSet() (ValidatorSet, error) {
	// This is a mock implementation of getValidatorSet
	// In a real implementation, this would query the Ethereum contract
	// to get the current validator set for the given epoch

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

func (e *ETHService) checkAndSubmitSignatures(ctx context.Context) error {
	return nil
}

func (e *ETHService) checkNewValidatorSet(ctx context.Context) error {

	// check set genesis

	phase, err := e.getCurrentPhase(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current phase: %w", err)
	}

	if phase != COMMIT {
		return nil
	}

	epoch, err := e.getCurrentEpoch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current epoch: %w", err)
	}

	timestamp, err := e.getCurrentEpochStart(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current epoch start: %w", err)
	}

	blockNumber, err := e.findBlockByTimestamp(ctx, timestamp)
	if err != nil {
		return fmt.Errorf("failed to find block by timestamp: %w", err)
	}

	valset, err := e.getValidatorSet(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to get validator set: %w", err)
	}

	// form valsetHeader, generate proofs

	// sign valsetHeader, proof

	// e.storage.AddSignature(epoch, "messageHash", "signature")

	return nil
}

func (e ETHService) getCurrentPhase(ctx context.Context) (Phase, error) {
	callMsg, err := constructCallMsg(e.contractAddress, e.contractABI, "getCurrentPhase")
	if err != nil {
		return 0, fmt.Errorf("failed to construct call msg: %w", err)
	}

	finalizedBlock, err := e.getFinalizedBlock()
	if err != nil {
		return 0, fmt.Errorf("failed to get finalized block: %w", err)
	}

	result, err := e.callContract(ctx, finalizedBlock, callMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	phase := new(big.Int).SetBytes(result).Uint64()
	return Phase(phase), nil
}

func (e ETHService) getCurrentEpoch(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.contractAddress, e.contractABI, "getCurrentEpoch")
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	finalizedBlock, err := e.getFinalizedBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get finalized block: %w", err)
	}

	result, err := e.callContract(ctx, finalizedBlock, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	epoch := new(big.Int).SetBytes(result)
	return epoch, nil
}

func (e ETHService) getCurrentEpochStart(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.contractAddress, e.contractABI, "getCurrentEpochStart")
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	finalizedBlock, err := e.getFinalizedBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get finalized block: %w", err)
	}

	result, err := e.callContract(ctx, finalizedBlock, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	epochStart := new(big.Int).SetBytes(result)
	return epochStart, nil
}

func (e ETHService) getEpochDuration(ctx context.Context) (*big.Int, error) {
	callMsg, err := constructCallMsg(e.contractAddress, e.contractABI, "getEpochDuration")
	if err != nil {
		return nil, fmt.Errorf("failed to construct call msg: %w", err)
	}

	finalizedBlock, err := e.getFinalizedBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get finalized block: %w", err)
	}

	result, err := e.callContract(ctx, finalizedBlock, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	epochDuration := new(big.Int).SetBytes(result)
	return epochDuration, nil
}

func (e *ETHService) getValidatorSet(ctx context.Context, blockNumber *big.Int) (ValidatorSet, error) {
	callMsg, err := constructCallMsg(e.contractAddress, e.contractABI, "getValidatorSet")
	if err != nil {
		return ValidatorSet{}, fmt.Errorf("failed to construct call msg: %w", err)
	}

	result, err := e.callContract(ctx, blockNumber, callMsg)
	if err != nil {
		return ValidatorSet{}, fmt.Errorf("failed to call contract: %w", err)
	}

	var valset ValidatorSet
	err = e.contractABI.UnpackIntoInterface(&valset, "getValidatorSet", result)
	if err != nil {
		return valset, err
	}

	return valset, nil
}

func (e ETHService) findBlockByTimestamp(ctx context.Context, timestamp *big.Int) (*big.Int, error) {
	// Get the latest block to use as upper bound
	latestBlock, err := e.client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	high := latestBlock.Number()
	low := big.NewInt(0)

	// Binary search to find the block closest to the target timestamp
	var mid *big.Int
	var closestBlock *big.Int
	var smallestDiff int64 = 1<<63 - 1 // Max int64 value

	for low.Cmp(high) <= 0 {
		mid = new(big.Int).Add(low, high)
		mid = mid.Div(mid, big.NewInt(2))

		// Get block at the middle
		block, err := e.client.BlockByNumber(ctx, mid)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %s: %v", mid.String(), err)
		}

		blockTime := int64(block.Time())
		diff := blockTime - timestamp.Int64()
		if diff < 0 {
			diff = -diff
		}

		// Update closest block if this is closer
		if diff < smallestDiff {
			smallestDiff = diff
			closestBlock = new(big.Int).Set(mid)
		}

		// If this block's timestamp is earlier than target, search higher
		if blockTime < timestamp.Int64() {
			low = new(big.Int).Add(mid, big.NewInt(1))
		} else if blockTime > timestamp.Int64() {
			// If this block's timestamp is later than target, search lower
			high = new(big.Int).Sub(mid, big.NewInt(1))
		} else {
			// Exact match
			return mid, nil
		}
	}

	return closestBlock, nil
}

func (e ETHService) getFinalizedBlock() (*big.Int, error) {
	// Get the latest finalized block number
	var result []byte
	err := e.client.Client().CallContext(context.Background(), &result, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalized block: %w", err)
	}

	blockNumber := new(big.Int).SetBytes(result)
	if blockNumber.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("failed to parse block number: %s", result)
	}

	return blockNumber, nil
}

func (e ETHService) callContract(ctx context.Context, blockNumber *big.Int, callMsg ethereum.CallMsg) ([]byte, error) {
	result, err := e.client.CallContract(ctx, callMsg, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return result, nil
}

func constructCallMsg(contractAddress common.Address, abi abi.ABI, method string, args ...interface{}) (ethereum.CallMsg, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return ethereum.CallMsg{}, fmt.Errorf("failed to pack method: %w", err)
	}

	return ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil
}
