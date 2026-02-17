package valsetDeriver

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/ssz"
	"github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver/mocks"
)

type testExternalVotingPowerClient struct {
	getVotingPowers func(ctx context.Context, address symbiotic.CrossChainAddress, timestamp symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error)
}

func (t *testExternalVotingPowerClient) GetVotingPowers(
	ctx context.Context,
	address symbiotic.CrossChainAddress,
	timestamp symbiotic.Timestamp,
) ([]symbiotic.OperatorVotingPower, error) {
	if t.getVotingPowers == nil {
		return nil, nil
	}
	return t.getVotingPowers(ctx, address, timestamp)
}

func TestDeriver_calcQuorumThreshold(t *testing.T) {
	tests := []struct {
		name           string
		config         symbiotic.NetworkConfig
		totalVP        symbiotic.VotingPower
		expectedQuorum *big.Int
		expectError    error
	}{
		{
			name: "valid quorum threshold calculation",
			config: symbiotic.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []symbiotic.QuorumThreshold{
					{
						KeyTag:          15,
						QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(670000000000000000)), // 67%
					},
				},
			},
			totalVP:        symbiotic.ToVotingPower(big.NewInt(1000)),
			expectedQuorum: big.NewInt(1000*.67 + 1), // (1000 * 67% + 1)
			expectError:    nil,
		},
		{
			name: "zero quorum threshold should error",
			config: symbiotic.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []symbiotic.QuorumThreshold{
					{
						KeyTag:          16,
						QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(670000000000000000)),
					},
				},
			},
			totalVP:     symbiotic.ToVotingPower(big.NewInt(1000)),
			expectError: errors.New("quorum threshold is zero"),
		},
		{
			name: "multiple thresholds - correct key selected",
			config: symbiotic.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []symbiotic.QuorumThreshold{
					{
						KeyTag:          16,
						QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(500000000000000000)),
					},
					{
						KeyTag:          15,
						QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(750000000000000000)), // 75%
					},
				},
			},
			totalVP:        symbiotic.ToVotingPower(big.NewInt(2000)),
			expectedQuorum: big.NewInt(2000*.75 + 1), // (2000 * 75% + 1)
			expectError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.CalcQuorumThreshold(tt.totalVP)

			if tt.expectError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedQuorum, result.Int)
			}
		})
	}
}

func TestDeriver_fillValidators(t *testing.T) {
	tests := []struct {
		name         string
		votingPowers []dtoOperatorVotingPower
		keys         []symbiotic.OperatorWithKeys
		expected     symbiotic.Validators
	}{
		{
			name: "single operator with voting power and keys",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
					},
				},
			},
			keys: []symbiotic.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
							ChainID:     1,
						},
					},
				},
			},
		},
		{
			name: "operator with multiple vaults and voting powers aggregated",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
								},
								{
									Vault:       common.HexToAddress("0x789"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
								},
							},
						},
					},
				},
			},
			keys: []symbiotic.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(800)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
							ChainID:     1,
						},
						{
							Vault:       common.HexToAddress("0x789"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
							ChainID:     1,
						},
					},
				},
			},
		},
		{
			name: "multiple operators",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
						{
							Operator: common.HexToAddress("0xabc"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0xdef"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
								},
							},
						},
					},
				},
			},
			keys: []symbiotic.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
				},
				{
					Operator: common.HexToAddress("0xabc"),
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(16),
							Payload: symbiotic.CompactPublicKey("key2"),
						},
					},
				},
			},
			expected: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
							ChainID:     1,
						},
					},
				},
				{
					Operator:    common.HexToAddress("0xabc"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(16),
							Payload: symbiotic.CompactPublicKey("key2"),
						},
					},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0xdef"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
							ChainID:     1,
						},
					},
				},
			},
		},
		{
			name: "operator with voting power but no keys",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
					},
				},
			},
			keys: []symbiotic.OperatorWithKeys{},
			expected: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys:        []symbiotic.ValidatorKey{},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
							ChainID:     1,
						},
					},
				},
			},
		},
		{
			name: "operator with multiple chains voting powers",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
								},
							},
						},
					},
				},
				{
					chainId: 137,
					votingPowers: []symbiotic.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []symbiotic.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x789"),
									VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
								},
							},
						},
					},
				},
			},
			keys: []symbiotic.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(800)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{
							Tag:     symbiotic.KeyTag(15),
							Payload: symbiotic.CompactPublicKey("key1"),
						},
					},
					Vaults: []symbiotic.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
							ChainID:     1,
						},
						{
							Vault:       common.HexToAddress("0x789"),
							VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
							ChainID:     137,
						},
					},
				},
			},
		},
		{
			name:         "empty inputs",
			votingPowers: []dtoOperatorVotingPower{},
			keys:         []symbiotic.OperatorWithKeys{},
			expected:     symbiotic.Validators{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillValidators(tt.votingPowers, tt.keys)

			require.Len(t, result, len(tt.expected))

			for i := range tt.expected {
				found := false
				for j := range result {
					if result[j].Operator == tt.expected[i].Operator {
						require.Equal(t, tt.expected[i].VotingPower.Int, result[j].VotingPower.Int)
						require.Equal(t, tt.expected[i].IsActive, result[j].IsActive)
						require.ElementsMatch(t, tt.expected[i].Keys, result[j].Keys)
						require.ElementsMatch(t, tt.expected[i].Vaults, result[j].Vaults)
						found = true
						break
					}
				}
				require.True(t, found, "expected validator with operator %s not found", tt.expected[i].Operator.Hex())
			}
		})
	}
}

func TestDeriver_fillValidatorsActive(t *testing.T) {
	tests := []struct {
		name                       string
		config                     symbiotic.NetworkConfig
		validators                 symbiotic.Validators
		expectedTotalVotingPower   *big.Int
		expectedActiveValidators   []common.Address    // operator addresses that should be active
		expectedInactiveValidators []common.Address    // operator addresses that should be inactive
		expectedVotingPowers       map[string]*big.Int // operator -> expected voting power after capping
	}{
		{
			name: "basic activation with minimum voting power",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(0)), // no max limit
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(0)), // no max count
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(50)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key3")},
					},
				},
			},
			expectedTotalVotingPower:   big.NewInt(700), // 500 + 200
			expectedActiveValidators:   []common.Address{common.HexToAddress("0x123"), common.HexToAddress("0x456")},
			expectedInactiveValidators: []common.Address{common.HexToAddress("0x789")}, // below minimum
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(500),
				"0x456": big.NewInt(200),
				"0x789": big.NewInt(50), // unchanged but inactive
			},
		},
		{
			name: "validator without keys should be inactive",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(0)),
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
					IsActive:    false,
					Keys:        []symbiotic.ValidatorKey{}, // no keys
				},
			},
			expectedTotalVotingPower:   big.NewInt(500),
			expectedActiveValidators:   []common.Address{common.HexToAddress("0x123")},
			expectedInactiveValidators: []common.Address{common.HexToAddress("0x456")}, // no keys
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(500),
				"0x456": big.NewInt(300),
			},
		},
		{
			name: "max voting power capping",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(400)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(0)),
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(600)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key2")},
					},
				},
			},
			expectedTotalVotingPower: big.NewInt(600), // 400 (capped) + 200
			expectedActiveValidators: []common.Address{common.HexToAddress("0x123"), common.HexToAddress("0x456")},
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(400), // capped from 600
				"0x456": big.NewInt(200),
			},
		},
		{
			name: "max validators count limit",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(2)),
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(400)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key3")},
					},
				},
			},
			expectedTotalVotingPower:   big.NewInt(900), // 500 + 400 (only first 2 validators)
			expectedActiveValidators:   []common.Address{common.HexToAddress("0x123"), common.HexToAddress("0x456")},
			expectedInactiveValidators: []common.Address{common.HexToAddress("0x789")}, // exceeds max count
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(500),
				"0x456": big.NewInt(400),
				"0x789": big.NewInt(300),
			},
		},
		{
			name: "combined constraints - min power, max power, max count",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(150)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(350)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(2)),
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(180)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key3")},
					},
				},
				{
					Operator:    common.HexToAddress("0xabc"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key4")},
					},
				},
			},
			expectedTotalVotingPower:   big.NewInt(550), // 350 (capped from 500) + 200
			expectedActiveValidators:   []common.Address{common.HexToAddress("0x123"), common.HexToAddress("0x456")},
			expectedInactiveValidators: []common.Address{common.HexToAddress("0x789"), common.HexToAddress("0xabc")}, // 0x789 exceeds max count, 0xabc below min power
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(350), // capped from 500
				"0x456": big.NewInt(200),
				"0x789": big.NewInt(180),
				"0xabc": big.NewInt(100),
			},
		},
		{
			name: "no validators meet criteria",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(0)),
			},
			validators: symbiotic.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []symbiotic.ValidatorKey{
						{Tag: symbiotic.KeyTag(15), Payload: symbiotic.CompactPublicKey("key1")},
					},
				},
			},
			expectedTotalVotingPower:   big.NewInt(0),
			expectedActiveValidators:   []common.Address{},
			expectedInactiveValidators: []common.Address{common.HexToAddress("0x123")},
			expectedVotingPowers: map[string]*big.Int{
				"0x123": big.NewInt(500),
			},
		},
		{
			name: "empty validators list",
			config: symbiotic.NetworkConfig{
				MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(0)),
			},
			validators:                 symbiotic.Validators{},
			expectedTotalVotingPower:   big.NewInt(0),
			expectedActiveValidators:   []common.Address{},
			expectedInactiveValidators: []common.Address{},
			expectedVotingPowers:       map[string]*big.Int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of validators to avoid modifying the test data
			validatorsCopy := make(symbiotic.Validators, len(tt.validators))
			for i, v := range tt.validators {
				validatorsCopy[i] = symbiotic.Validator{
					Operator:    v.Operator,
					VotingPower: symbiotic.ToVotingPower(new(big.Int).Set(v.VotingPower.Int)),
					IsActive:    v.IsActive,
					Keys:        append([]symbiotic.ValidatorKey{}, v.Keys...),
					Vaults:      append([]symbiotic.ValidatorVault{}, v.Vaults...),
				}
			}

			markValidatorsActive(tt.config, validatorsCopy)

			// Check total voting power
			totalVotingPower := validatorsCopy.GetTotalActiveVotingPower()
			require.Equal(t, tt.expectedTotalVotingPower, totalVotingPower.Int, "total voting power mismatch")

			// Check active validators
			var activeValidators []common.Address
			var inactiveValidators []common.Address
			for _, validator := range validatorsCopy {
				addr := validator.Operator.Hex()
				if validator.IsActive {
					activeValidators = append(activeValidators, validator.Operator)
				} else {
					inactiveValidators = append(inactiveValidators, validator.Operator)
				}

				// Check voting power (capping)
				if expectedVP, exists := tt.expectedVotingPowers[addr]; exists {
					require.Equal(t, expectedVP, validator.VotingPower.Int,
						"voting power mismatch for validator %s", addr)
				}
			}

			require.ElementsMatch(t, tt.expectedActiveValidators, activeValidators,
				"active validators mismatch")
			require.ElementsMatch(t, tt.expectedInactiveValidators, inactiveValidators,
				"inactive validators mismatch")
		})
	}
}

func TestDeriver_GetNetworkData(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(evmClient *mocks.MockEvmClient)
		addr       symbiotic.CrossChainAddress
		expected   symbiotic.NetworkData
		errorMsg   string
	}{
		{
			name: "successful get network data",
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.HexToHash("0x456"), nil)
				m.EXPECT().GetEip712Domain(gomock.Any(), gomock.Any()).Return(symbiotic.Eip712Domain{
					Fields:            [1]byte{},
					Name:              "TestNetwork",
					Version:           "1",
					ChainId:           big.NewInt(111),
					VerifyingContract: common.HexToAddress("0x123"),
					Salt:              [32]byte{1, 2, 3},
					Extensions:        []*big.Int{big.NewInt(333)},
				}, nil)
			},
			addr: symbiotic.CrossChainAddress{},
			expected: symbiotic.NetworkData{
				Address:    common.HexToAddress("0x123"),
				Subnetwork: common.HexToHash("0x456"),
				Eip712Data: symbiotic.Eip712Domain{
					Fields:            [1]byte{},
					Name:              "TestNetwork",
					Version:           "1",
					ChainId:           big.NewInt(111),
					VerifyingContract: common.HexToAddress("0x123"),
					Salt:              [32]byte{1, 2, 3},
					Extensions:        []*big.Int{big.NewInt(333)},
				},
			},
		},
		{
			name: "network address error",
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.Address{}, errors.New("network address error"))
			},
			addr:     symbiotic.CrossChainAddress{},
			errorMsg: "failed to get network address",
		},
		{
			name: "subnetwork error",
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.Hash{}, errors.New("subnetwork error"))
			},
			addr:     symbiotic.CrossChainAddress{},
			errorMsg: "failed to get subnetwork",
		},
		{
			name: "eip712 domain error",
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.HexToHash("0x456"), nil)
				m.EXPECT().GetEip712Domain(gomock.Any(), gomock.Any()).Return(symbiotic.Eip712Domain{}, errors.New("eip712 error"))
			},
			addr:     symbiotic.CrossChainAddress{},
			errorMsg: "failed to get eip712 domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEvmClient := mocks.NewMockEvmClient(ctrl)
			tt.setupMocks(mockEvmClient)

			d, err := NewDeriver(mockEvmClient, nil)
			require.NoError(t, err)

			result, err := d.GetNetworkData(context.Background(), tt.addr)

			if tt.errorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDeriver_fillValidators_VaultLimitExceeded(t *testing.T) {
	// This test verifies vault truncation when exceeding ssz.VaultsListMaxElements (1024)
	// Expected behavior:
	// 1. Sort vaults by voting power DESC, then vault address ASC
	// 2. Truncate to keep only top 1024 vaults
	// 3. Re-sort remaining vaults by vault address ASC only

	const (
		highPowerVaults   = 5
		mediumPowerVaults = 5
		extraVaults       = 10 // Create more than the 1024 limit
		highPower         = 2000
		mediumPower       = 1500
		lowPower          = 1000
	)

	totalVaults := ssz.VaultsListMaxElements + extraVaults
	vaults := make([]symbiotic.VaultVotingPower, totalVaults)

	// Create test vaults with different voting powers
	for i := 0; i < totalVaults; i++ {
		var power int64
		switch {
		case i < highPowerVaults:
			power = highPower
		case i < highPowerVaults+mediumPowerVaults:
			power = mediumPower
		default:
			power = lowPower
		}

		vaults[i] = symbiotic.VaultVotingPower{
			Vault:       common.HexToAddress(fmt.Sprintf("0x%040d", i+1)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(power)),
		}
	}

	// Setup test data
	votingPowers := []dtoOperatorVotingPower{{
		chainId: 1,
		votingPowers: []symbiotic.OperatorVotingPower{{
			Operator: common.HexToAddress("0x123"),
			Vaults:   vaults,
		}},
	}}

	keys := []symbiotic.OperatorWithKeys{{
		Operator: common.HexToAddress("0x123"),
		Keys: []symbiotic.ValidatorKey{{
			Tag:     symbiotic.KeyTag(15),
			Payload: symbiotic.CompactPublicKey("key1"),
		}},
	}}

	// Execute
	result := fillValidators(votingPowers, keys)

	// Verify results
	require.Len(t, result, 1)
	validator := result[0]

	t.Run("vault count is limited", func(t *testing.T) {
		require.Len(t, validator.Vaults, ssz.VaultsListMaxElements)
	})

	t.Run("total voting power matches kept vaults", func(t *testing.T) {
		expectedTotal := big.NewInt(0)
		for _, vault := range validator.Vaults {
			expectedTotal.Add(expectedTotal, vault.VotingPower.Int)
		}
		require.Equal(t, expectedTotal, validator.VotingPower.Int)
	})

	t.Run("vaults are sorted by address ascending", func(t *testing.T) {
		for i := 1; i < len(validator.Vaults); i++ {
			prev := validator.Vaults[i-1].Vault.Hex()
			curr := validator.Vaults[i].Vault.Hex()
			require.Less(t, prev, curr,
				"Vault %s at index %d should come before vault %s at index %d",
				prev, i-1, curr, i)
		}
	})

	t.Run("highest power vaults are retained", func(t *testing.T) {
		powerCounts := map[int64]int{}
		for _, vault := range validator.Vaults {
			power := vault.VotingPower.Int64()
			powerCounts[power]++
		}

		// All high and medium power vaults should be kept
		require.Equal(t, highPowerVaults, powerCounts[highPower],
			"Should keep all %d high power vaults", highPowerVaults)
		require.Equal(t, mediumPowerVaults, powerCounts[mediumPower],
			"Should keep all %d medium power vaults", mediumPowerVaults)

		// Remaining slots filled with low power vaults
		expectedLowPowerCount := ssz.VaultsListMaxElements - highPowerVaults - mediumPowerVaults
		require.Equal(t, expectedLowPowerCount, powerCounts[lowPower],
			"Should keep %d low power vaults", expectedLowPowerCount)
	})

	t.Run("validator properties are correct", func(t *testing.T) {
		require.Equal(t, common.HexToAddress("0x123"), validator.Operator)
		require.False(t, validator.IsActive)
		require.Len(t, validator.Keys, 1)
		require.Equal(t, symbiotic.KeyTag(15), validator.Keys[0].Tag)
	})
}

func TestDeriver_getVotingPowersFromProviders_MixedRouting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockEvmClient(ctrl)
	timestamp := symbiotic.Timestamp(123)
	evmProvider := symbiotic.CrossChainAddress{
		ChainId: 1,
		Address: common.HexToAddress("0x0000000000000000000000000000000000000011"),
	}
	externalProvider := symbiotic.CrossChainAddress{
		ChainId: 4_000_000_001,
		Address: common.HexToAddress("0x1122334455667788990000000000000000000000"),
	}

	mockEvmClient.EXPECT().
		GetVotingPowers(gomock.Any(), evmProvider, timestamp).
		Return([]symbiotic.OperatorVotingPower{
			{
				Operator: common.HexToAddress("0x0000000000000000000000000000000000000001"),
				Vaults: []symbiotic.VaultVotingPower{
					{
						Vault:       common.HexToAddress("0x0000000000000000000000000000000000000011"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(10)),
					},
				},
			},
		}, nil)

	externalCalled := false
	externalClient := &testExternalVotingPowerClient{
		getVotingPowers: func(_ context.Context, address symbiotic.CrossChainAddress, _ symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error) {
			externalCalled = true
			require.Equal(t, externalProvider, address)
			return []symbiotic.OperatorVotingPower{
				{
					Operator: common.HexToAddress("0x0000000000000000000000000000000000000002"),
					Vaults: []symbiotic.VaultVotingPower{
						{
							Vault:       externalProvider.Address,
							VotingPower: symbiotic.ToVotingPower(big.NewInt(20)),
						},
					},
				},
			}, nil
		},
	}

	d, err := NewDeriver(mockEvmClient, externalClient)
	require.NoError(t, err)

	result, err := d.getVotingPowersFromProviders(context.Background(), []symbiotic.CrossChainAddress{evmProvider, externalProvider}, timestamp)
	require.NoError(t, err)
	require.True(t, externalCalled)
	require.Len(t, result, 2)
	require.Equal(t, uint64(1), result[0].chainId)
	require.Equal(t, uint64(4_000_000_001), result[1].chainId)
}

func TestDeriver_getVotingPowersFromProviders_ChainIDZeroUsesEVM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEvmClient := mocks.NewMockEvmClient(ctrl)
	timestamp := symbiotic.Timestamp(123)
	provider := symbiotic.CrossChainAddress{
		ChainId: 0,
		Address: common.HexToAddress("0x0000000000000000000000000000000000000011"),
	}

	mockEvmClient.EXPECT().
		GetVotingPowers(gomock.Any(), provider, timestamp).
		Return([]symbiotic.OperatorVotingPower{}, nil)

	externalCalled := false
	externalClient := &testExternalVotingPowerClient{
		getVotingPowers: func(_ context.Context, _ symbiotic.CrossChainAddress, _ symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error) {
			externalCalled = true
			return []symbiotic.OperatorVotingPower{}, nil
		},
	}

	d, err := NewDeriver(mockEvmClient, externalClient)
	require.NoError(t, err)

	_, err = d.getVotingPowersFromProviders(context.Background(), []symbiotic.CrossChainAddress{provider}, timestamp)
	require.NoError(t, err)
	require.False(t, externalCalled)
}

func TestDeriver_getVotingPowersFromProviders_NoExternalClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d, err := NewDeriver(mocks.NewMockEvmClient(ctrl), nil)
	require.NoError(t, err)

	_, err = d.getVotingPowersFromProviders(context.Background(), []symbiotic.CrossChainAddress{{
		ChainId: 4_000_000_001,
		Address: common.HexToAddress("0x1122334455667788990000000000000000000000"),
	}}, symbiotic.Timestamp(1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "external voting power client is not configured")
}

func TestDeriver_getVotingPowersFromProviders_TypedNilExternalClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var typedNilExternalClient *testExternalVotingPowerClient

	d, err := NewDeriver(mocks.NewMockEvmClient(ctrl), typedNilExternalClient)
	require.NoError(t, err)

	var gotErr error
	require.NotPanics(t, func() {
		_, gotErr = d.getVotingPowersFromProviders(context.Background(), []symbiotic.CrossChainAddress{{
			ChainId: 4_000_000_001,
			Address: common.HexToAddress("0x1122334455667788990000000000000000000000"),
		}}, symbiotic.Timestamp(1))
	})
	require.Error(t, gotErr)
	require.Contains(t, gotErr.Error(), "external voting power client is not configured")
}

func TestDeriver_getVotingPowersFromProviders_ErrorPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	externalClient := &testExternalVotingPowerClient{
		getVotingPowers: func(_ context.Context, _ symbiotic.CrossChainAddress, _ symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error) {
			return nil, errors.New("external provider failure")
		},
	}

	d, err := NewDeriver(mocks.NewMockEvmClient(ctrl), externalClient)
	require.NoError(t, err)

	_, err = d.getVotingPowersFromProviders(context.Background(), []symbiotic.CrossChainAddress{{
		ChainId: 4_000_000_001,
		Address: common.HexToAddress("0x1122334455667788990000000000000000000000"),
	}}, symbiotic.Timestamp(1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "external provider failure")
}

func TestDeriver_getVotingPowersFromProviders_ConcurrencyLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var inFlight int64
	var maxInFlight int64

	externalClient := &testExternalVotingPowerClient{
		getVotingPowers: func(_ context.Context, _ symbiotic.CrossChainAddress, _ symbiotic.Timestamp) ([]symbiotic.OperatorVotingPower, error) {
			current := atomic.AddInt64(&inFlight, 1)
			for {
				prev := atomic.LoadInt64(&maxInFlight)
				if current <= prev {
					break
				}
				if atomic.CompareAndSwapInt64(&maxInFlight, prev, current) {
					break
				}
			}

			time.Sleep(40 * time.Millisecond)
			atomic.AddInt64(&inFlight, -1)
			return []symbiotic.OperatorVotingPower{}, nil
		},
	}

	d, err := NewDeriver(mocks.NewMockEvmClient(ctrl), externalClient)
	require.NoError(t, err)

	providers := make([]symbiotic.CrossChainAddress, 25)
	for i := 0; i < len(providers); i++ {
		providers[i] = symbiotic.CrossChainAddress{
			ChainId: 4_000_000_001 + uint64(i),
			Address: common.HexToAddress(fmt.Sprintf("0x%040x", i+1)),
		}
	}

	_, err = d.getVotingPowersFromProviders(context.Background(), providers, symbiotic.Timestamp(1))
	require.NoError(t, err)
	require.LessOrEqual(t, maxInFlight, int64(10))
	require.Greater(t, maxInFlight, int64(1))
}

func TestDeriver_GetSchedulerInfo(t *testing.T) {
	tests := []struct {
		name                string
		valset              symbiotic.ValidatorSet
		config              symbiotic.NetworkConfig
		expectedAggIndices  []uint32
		expectedCommIndices []uint32
		expectError         bool
		errorMsg            string
	}{
		{
			name: "basic scheduling with 3 validators, 2 aggregators, 1 committer",
			valset: symbiotic.ValidatorSet{
				Validators: symbiotic.Validators{
					{
						Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
						IsActive:    true,
					},
					{
						Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
						IsActive:    true,
					},
					{
						Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(1500)),
						IsActive:    true,
					},
				},
				Version:          1,
				RequiredKeyTag:   15,
				Epoch:            100,
				CaptureTimestamp: 1234567890,
			},
			config: symbiotic.NetworkConfig{
				NumAggregators: 2,
				NumCommitters:  1,
			},
			// These expected values are deterministic based on the hash calculation
			expectedAggIndices:  []uint32{2, 0}, // Calculated deterministically from hash
			expectedCommIndices: []uint32{1},    // Calculated deterministically from hash
			expectError:         false,
		},
		{
			name: "single validator with multiple roles",
			valset: symbiotic.ValidatorSet{
				Validators: symbiotic.Validators{
					{
						Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
						IsActive:    true,
					},
				},
				Version:          1,
				RequiredKeyTag:   15,
				Epoch:            50,
				CaptureTimestamp: 1234567890,
			},
			config: symbiotic.NetworkConfig{
				NumAggregators: 1,
				NumCommitters:  1,
			},
			expectedAggIndices:  []uint32{0},
			expectedCommIndices: []uint32{0},
			expectError:         false,
		},
		{
			name: "zero aggregators and committers",
			valset: symbiotic.ValidatorSet{
				Validators: symbiotic.Validators{
					{
						Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
						IsActive:    true,
					},
				},
				Version:          1,
				RequiredKeyTag:   15,
				Epoch:            50,
				CaptureTimestamp: 1234567890,
			},
			config: symbiotic.NetworkConfig{
				NumAggregators: 0,
				NumCommitters:  0,
			},
			expectedAggIndices:  nil,
			expectedCommIndices: nil,
			expectError:         false,
		},
		{
			name: "empty validator set",
			valset: symbiotic.ValidatorSet{
				Validators:       symbiotic.Validators{},
				Version:          1,
				RequiredKeyTag:   15,
				Epoch:            100,
				CaptureTimestamp: 1234567890,
			},
			config: symbiotic.NetworkConfig{
				NumAggregators: 2,
				NumCommitters:  1,
			},
			expectedAggIndices:  []uint32{},
			expectedCommIndices: []uint32{},
			expectError:         false,
		},
		{
			name: "unsorted validators should error",
			valset: symbiotic.ValidatorSet{
				Validators: symbiotic.Validators{
					{
						Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
						IsActive:    true,
					},
					{
						Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
						IsActive:    true,
					},
				},
				Version:          1,
				RequiredKeyTag:   15,
				Epoch:            100,
				CaptureTimestamp: 1234567890,
			},
			config: symbiotic.NetworkConfig{
				NumAggregators: 1,
				NumCommitters:  1,
			},
			expectError: true,
			errorMsg:    "validators are not sorted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggIndices, commIndices, err := GetSchedulerInfo(context.Background(), tt.valset, tt.config)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)

			require.ElementsMatch(t, tt.expectedAggIndices, aggIndices)
			require.ElementsMatch(t, tt.expectedCommIndices, commIndices)
		})
	}
}

func TestDeriver_GetSchedulerInfo_Deterministic(t *testing.T) {
	// Test that GetSchedulerInfo returns consistent results for the same inputs
	valset := symbiotic.ValidatorSet{
		Validators: symbiotic.Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1500)),
				IsActive:    true,
			},
		},
		Version:          1,
		RequiredKeyTag:   15,
		Epoch:            100,
		CaptureTimestamp: 1234567890,
	}

	config := symbiotic.NetworkConfig{
		NumAggregators: 2,
		NumCommitters:  1,
	}

	// Run the same calculation multiple times
	const iterations = 10
	var firstAggIndices, firstCommIndices []uint32

	for i := 0; i < iterations; i++ {
		aggIndices, commIndices, err := GetSchedulerInfo(context.Background(), valset, config)
		require.NoError(t, err)

		if i == 0 {
			firstAggIndices = aggIndices
			firstCommIndices = commIndices
		} else {
			require.ElementsMatch(t, firstAggIndices, aggIndices,
				"aggregator indices should be deterministic, iteration %d", i)
			require.ElementsMatch(t, firstCommIndices, commIndices,
				"committer indices should be deterministic, iteration %d", i)
		}
	}
}

func TestDeriver_GetSchedulerInfo_VerifyRandomness(t *testing.T) {
	// Test that different inputs produce different results
	baseValset := symbiotic.ValidatorSet{
		Validators: symbiotic.Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(2000)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1500)),
				IsActive:    true,
			},
		},
		Version:          1,
		RequiredKeyTag:   15,
		Epoch:            100,
		CaptureTimestamp: 1234567890,
	}

	config := symbiotic.NetworkConfig{
		NumAggregators: 2,
		NumCommitters:  1,
	}

	// Get results for original valset
	aggIndices1, commIndices1, err := GetSchedulerInfo(context.Background(), baseValset, config)
	require.NoError(t, err)

	// Test with different epoch
	valsetDifferentEpoch := baseValset
	valsetDifferentEpoch.Epoch = 101
	aggIndices2, commIndices2, err := GetSchedulerInfo(context.Background(), valsetDifferentEpoch, config)
	require.NoError(t, err)

	// Note: Different epochs might produce the same results due to hash collisions,
	// but we can verify they run without error
	_ = aggIndices2
	_ = commIndices2

	// Test with different timestamp
	valsetDifferentTimestamp := baseValset
	valsetDifferentTimestamp.CaptureTimestamp = 9876543210
	aggIndices3, commIndices3, err := GetSchedulerInfo(context.Background(), valsetDifferentTimestamp, config)
	require.NoError(t, err)

	// Results should be different for different timestamps (high probability)
	differentAgg := !slices.Equal(aggIndices1, aggIndices3)
	differentComm := !slices.Equal(commIndices1, commIndices3)
	require.True(t, differentAgg || differentComm,
		"different timestamps should likely produce different results")
}

func TestDeriver_findNextAvailableIndex(t *testing.T) {
	tests := []struct {
		name           string
		startIndex     uint32
		validatorCount int
		usedIndices    map[uint32]struct{}
		expected       uint32
	}{
		{
			name:           "first index available",
			startIndex:     0,
			validatorCount: 5,
			usedIndices:    map[uint32]struct{}{},
			expected:       0,
		},
		{
			name:           "start index taken, next available",
			startIndex:     2,
			validatorCount: 5,
			usedIndices:    map[uint32]struct{}{2: {}},
			expected:       3,
		},
		{
			name:           "wrap around to beginning",
			startIndex:     4,
			validatorCount: 5,
			usedIndices:    map[uint32]struct{}{4: {}},
			expected:       0,
		},
		{
			name:           "multiple indices taken, find next available",
			startIndex:     1,
			validatorCount: 5,
			usedIndices:    map[uint32]struct{}{1: {}, 2: {}, 3: {}},
			expected:       4,
		},
		{
			name:           "wrap around with multiple taken",
			startIndex:     3,
			validatorCount: 4,
			usedIndices:    map[uint32]struct{}{3: {}, 0: {}},
			expected:       1,
		},
		{
			name:           "single validator, not taken",
			startIndex:     0,
			validatorCount: 1,
			usedIndices:    map[uint32]struct{}{},
			expected:       0,
		},
		{
			name:           "large validator set, find available",
			startIndex:     100,
			validatorCount: 200,
			usedIndices:    map[uint32]struct{}{100: {}, 101: {}, 102: {}},
			expected:       103,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findNextAvailableIndex(tt.startIndex, tt.validatorCount, tt.usedIndices)
			require.Equal(t, tt.expected, result)

			// Verify the result is not in usedIndices
			_, exists := tt.usedIndices[result]
			require.False(t, exists, "returned index should not be in used indices")

			// Verify the result is within bounds
			require.Less(t, result, uint32(tt.validatorCount))
		})
	}
}

func TestDeriver_findNextAvailableIndex_Panic(t *testing.T) {
	// Test that the function panics when no indices are available

	// Create a scenario where all indices are taken
	usedIndices := map[uint32]struct{}{
		0: {}, 1: {}, 2: {},
	}

	require.Panics(t, func() {
		findNextAvailableIndex(0, 3, usedIndices)
	}, "should panic when no indices are available")
}
