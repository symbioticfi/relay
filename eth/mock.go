package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// MockEthClient implements EthClient interface for testing
type MockEthClient struct {
	// Mock return values
	MockValidatorSet      ValidatorSet
	MockCurrentPhase      Phase
	MockCurrentEpoch      *big.Int
	MockCurrentEpochStart *big.Int
	MockIsGenesisSet      bool
	MockEpochDuration     *big.Int
	MockRequiredKeyTag    uint8
	MockQuorumThreshold   *big.Int
	MockFinalizedBlock    *big.Int
	MockCommitError       error
	MockBlockForTimestamp *big.Int
}

// NewMockEthClient creates a new mock client with default values
func NewMockEthClient() *MockEthClient {
	mockValidators := createMockValidators(5)
	totalVotingPower := big.NewInt(0)
	for _, v := range mockValidators {
		if v.IsActive {
			totalVotingPower.Add(totalVotingPower, v.VotingPower)
		}
	}

	return &MockEthClient{
		MockValidatorSet: ValidatorSet{
			Version:                1,
			TotalActiveVotingPower: totalVotingPower,
			Validators:             mockValidators,
		},
		MockCurrentPhase:      COMMIT,
		MockCurrentEpoch:      big.NewInt(1),
		MockCurrentEpochStart: big.NewInt(1000),
		MockIsGenesisSet:      true,
		MockEpochDuration:     big.NewInt(100),
		MockRequiredKeyTag:    1,
		MockQuorumThreshold:   big.NewInt(667), // 2/3 of 1000
		MockFinalizedBlock:    big.NewInt(12345),
		MockBlockForTimestamp: big.NewInt(12345),
	}
}

// createMockValidators creates a list of mock validators for testing
func createMockValidators(count int) []*Validator {
	validators := make([]*Validator, count)

	for i := 0; i < count; i++ {
		// Create BLS key
		blsKey := &Key{
			Tag:     1, // BLS key tag
			Payload: make([]byte, 48),
		}
		// Just put some dummy data in the payload
		for j := 0; j < len(blsKey.Payload); j++ {
			blsKey.Payload[j] = byte(i + j)
		}

		// Create ECDSA key
		ecdsaKey := &Key{
			Tag:     2, // ECDSA key tag
			Payload: make([]byte, 33),
		}
		// Just put some dummy data in the payload
		for j := 0; j < len(ecdsaKey.Payload); j++ {
			ecdsaKey.Payload[j] = byte(i + j + 100)
		}

		// Create a vault
		vault := &Vault{
			VaultAddress: common.HexToAddress(generateMockAddress(i)),
			VotingPower:  big.NewInt(int64(200 + i*10)),
		}

		validators[i] = &Validator{
			Version:     1,
			Operator:    common.HexToAddress(generateMockAddress(i + 100)),
			VotingPower: big.NewInt(int64(200 + i*10)),
			IsActive:    true,
			Keys:        []*Key{blsKey, ecdsaKey},
			Vaults:      []*Vault{vault},
		}
	}

	return validators
}

// generateMockAddress creates a mock Ethereum address for testing
func generateMockAddress(seed int) string {
	return "0x" + padLeft(intToHex(seed), 40, '0')
}

// intToHex converts an integer to a hex string
func intToHex(n int) string {
	const hexChars = "0123456789abcdef"
	if n == 0 {
		return "0"
	}

	var result string
	for n > 0 {
		result = string(hexChars[n%16]) + result
		n /= 16
	}
	return result
}

// padLeft pads a string with a character to a specific length
func padLeft(str string, length int, pad byte) string {
	if len(str) >= length {
		return str
	}

	padding := make([]byte, length-len(str))
	for i := range padding {
		padding[i] = pad
	}

	return string(padding) + str
}

// Commit mocks the Commit method
func (m *MockEthClient) Commit(messageHash string, signature []byte) error {
	return nil
}

// GetNewValidatorSet mocks the GetNewValidatorSet method
func (m *MockEthClient) GetNewValidatorSet(ctx context.Context) (*ValidatorSet, error) {
	return &m.MockValidatorSet, nil
}

// GetCurrentPhase mocks the GetCurrentPhase method
func (m *MockEthClient) GetCurrentPhase(ctx context.Context) (Phase, error) {
	return m.MockCurrentPhase, nil
}

// GetCurrentEpoch mocks the GetCurrentEpoch method
func (m *MockEthClient) GetCurrentEpoch(ctx context.Context) (*big.Int, error) {
	return m.MockCurrentEpoch, nil
}

// GetCurrentEpochStart mocks the GetCurrentEpochStart method
func (m *MockEthClient) GetCurrentEpochStart(ctx context.Context) (*big.Int, error) {
	return m.MockCurrentEpochStart, nil
}

// GetIsGenesisSet mocks the GetIsGenesisSet method
func (m *MockEthClient) GetIsGenesisSet(ctx context.Context) (bool, error) {
	return m.MockIsGenesisSet, nil
}

// GetEpochDuration mocks the GetEpochDuration method
func (m *MockEthClient) GetEpochDuration(ctx context.Context) (*big.Int, error) {
	return m.MockEpochDuration, nil
}

// GetRequiredKeyTag mocks the GetRequiredKeyTag method
func (m *MockEthClient) GetRequiredKeyTag(ctx context.Context) (uint8, error) {
	return m.MockRequiredKeyTag, nil
}

// GetQuorumThreshold mocks the GetQuorumThreshold method
func (m *MockEthClient) GetQuorumThreshold(ctx context.Context) (*big.Int, error) {
	return m.MockQuorumThreshold, nil
}

// GetValidatorSet mocks the GetValidatorSet method
func (m *MockEthClient) GetValidatorSet(ctx context.Context, blockNumber *big.Int) (ValidatorSet, error) {
	return m.MockValidatorSet, nil
}

// GetFinalizedBlock mocks the GetFinalizedBlock method
func (m *MockEthClient) GetFinalizedBlock() (*big.Int, error) {
	return m.MockFinalizedBlock, nil
}

// findBlockByTimestamp mocks the findBlockByTimestamp method
func (m *MockEthClient) findBlockByTimestamp(ctx context.Context, timestamp *big.Int) (*big.Int, error) {
	return m.MockBlockForTimestamp, nil
}
