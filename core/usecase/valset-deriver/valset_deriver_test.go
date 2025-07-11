package valsetDeriver

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/valset-deriver/mocks"
)

func TestDeriver_calcQuorumThreshold(t *testing.T) {
	tests := []struct {
		name           string
		config         entity.NetworkConfig
		totalVP        entity.VotingPower
		expectedQuorum *big.Int
		expectError    error
	}{
		{
			name: "valid quorum threshold calculation",
			config: entity.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []entity.QuorumThreshold{
					{
						KeyTag:          15,
						QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(670000000000000000)), // 67%
					},
				},
			},
			totalVP:        entity.ToVotingPower(big.NewInt(1000)),
			expectedQuorum: big.NewInt(1000*.67 + 1), // (1000 * 67% + 1)
			expectError:    nil,
		},
		{
			name: "zero quorum threshold should error",
			config: entity.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []entity.QuorumThreshold{
					{
						KeyTag:          16,
						QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(670000000000000000)),
					},
				},
			},
			totalVP:     entity.ToVotingPower(big.NewInt(1000)),
			expectError: errors.New("quorum threshold is zero"),
		},
		{
			name: "multiple thresholds - correct key selected",
			config: entity.NetworkConfig{
				RequiredHeaderKeyTag: 15,
				QuorumThresholds: []entity.QuorumThreshold{
					{
						KeyTag:          16,
						QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(500000000000000000)),
					},
					{
						KeyTag:          15,
						QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(750000000000000000)), // 75%
					},
				},
			},
			totalVP:        entity.ToVotingPower(big.NewInt(2000)),
			expectedQuorum: big.NewInt(2000*.75 + 1), // (2000 * 75% + 1)
			expectError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDeriver(nil)
			require.NoError(t, err)
			result, err := d.calcQuorumThreshold(tt.config, tt.totalVP)

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
		keys         []entity.OperatorWithKeys
		expected     entity.Validators
	}{
		{
			name: "single operator with voting power and keys",
			votingPowers: []dtoOperatorVotingPower{
				{
					chainId: 1,
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: entity.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
					},
				},
			},
			keys: []entity.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: entity.ToVotingPower(big.NewInt(1000)),
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
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: entity.ToVotingPower(big.NewInt(500)),
								},
								{
									Vault:       common.HexToAddress("0x789"),
									VotingPower: entity.ToVotingPower(big.NewInt(300)),
								},
							},
						},
					},
				},
			},
			keys: []entity.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(800)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: entity.ToVotingPower(big.NewInt(500)),
							ChainID:     1,
						},
						{
							Vault:       common.HexToAddress("0x789"),
							VotingPower: entity.ToVotingPower(big.NewInt(300)),
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
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: entity.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
						{
							Operator: common.HexToAddress("0xabc"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0xdef"),
									VotingPower: entity.ToVotingPower(big.NewInt(2000)),
								},
							},
						},
					},
				},
			},
			keys: []entity.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
				},
				{
					Operator: common.HexToAddress("0xabc"),
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(16),
							Payload: entity.CompactPublicKey("key2"),
						},
					},
				},
			},
			expected: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: entity.ToVotingPower(big.NewInt(1000)),
							ChainID:     1,
						},
					},
				},
				{
					Operator:    common.HexToAddress("0xabc"),
					VotingPower: entity.ToVotingPower(big.NewInt(2000)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(16),
							Payload: entity.CompactPublicKey("key2"),
						},
					},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0xdef"),
							VotingPower: entity.ToVotingPower(big.NewInt(2000)),
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
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: entity.ToVotingPower(big.NewInt(1000)),
								},
							},
						},
					},
				},
			},
			keys: []entity.OperatorWithKeys{},
			expected: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(1000)),
					IsActive:    false,
					Keys:        []entity.ValidatorKey{},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: entity.ToVotingPower(big.NewInt(1000)),
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
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x456"),
									VotingPower: entity.ToVotingPower(big.NewInt(500)),
								},
							},
						},
					},
				},
				{
					chainId: 137,
					votingPowers: []entity.OperatorVotingPower{
						{
							Operator: common.HexToAddress("0x123"),
							Vaults: []entity.VaultVotingPower{
								{
									Vault:       common.HexToAddress("0x789"),
									VotingPower: entity.ToVotingPower(big.NewInt(300)),
								},
							},
						},
					},
				},
			},
			keys: []entity.OperatorWithKeys{
				{
					Operator: common.HexToAddress("0x123"),
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
				},
			},
			expected: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(800)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{
							Tag:     entity.KeyTag(15),
							Payload: entity.CompactPublicKey("key1"),
						},
					},
					Vaults: []entity.ValidatorVault{
						{
							Vault:       common.HexToAddress("0x456"),
							VotingPower: entity.ToVotingPower(big.NewInt(500)),
							ChainID:     1,
						},
						{
							Vault:       common.HexToAddress("0x789"),
							VotingPower: entity.ToVotingPower(big.NewInt(300)),
							ChainID:     137,
						},
					},
				},
			},
		},
		{
			name:         "empty inputs",
			votingPowers: []dtoOperatorVotingPower{},
			keys:         []entity.OperatorWithKeys{},
			expected:     entity.Validators{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDeriver(nil)
			require.NoError(t, err)

			result := d.fillValidators(tt.votingPowers, tt.keys)

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
		config                     entity.NetworkConfig
		validators                 entity.Validators
		expectedTotalVotingPower   *big.Int
		expectedActiveValidators   []common.Address    // operator addresses that should be active
		expectedInactiveValidators []common.Address    // operator addresses that should be inactive
		expectedVotingPowers       map[string]*big.Int // operator -> expected voting power after capping
	}{
		{
			name: "basic activation with minimum voting power",
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)), // no max limit
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)), // no max count
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: entity.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: entity.ToVotingPower(big.NewInt(50)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key3")},
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)),
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: entity.ToVotingPower(big.NewInt(300)),
					IsActive:    false,
					Keys:        []entity.ValidatorKey{}, // no keys
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(400)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)),
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(600)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: entity.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key2")},
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(2)),
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: entity.ToVotingPower(big.NewInt(400)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: entity.ToVotingPower(big.NewInt(300)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key3")},
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(150)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(350)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(2)),
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
					},
				},
				{
					Operator:    common.HexToAddress("0x456"),
					VotingPower: entity.ToVotingPower(big.NewInt(200)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key2")},
					},
				},
				{
					Operator:    common.HexToAddress("0x789"),
					VotingPower: entity.ToVotingPower(big.NewInt(180)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key3")},
					},
				},
				{
					Operator:    common.HexToAddress("0xabc"),
					VotingPower: entity.ToVotingPower(big.NewInt(100)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key4")},
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(1000)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)),
			},
			validators: entity.Validators{
				{
					Operator:    common.HexToAddress("0x123"),
					VotingPower: entity.ToVotingPower(big.NewInt(500)),
					IsActive:    false,
					Keys: []entity.ValidatorKey{
						{Tag: entity.KeyTag(15), Payload: entity.CompactPublicKey("key1")},
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
			config: entity.NetworkConfig{
				MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
				MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)),
				MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)),
			},
			validators:                 entity.Validators{},
			expectedTotalVotingPower:   big.NewInt(0),
			expectedActiveValidators:   []common.Address{},
			expectedInactiveValidators: []common.Address{},
			expectedVotingPowers:       map[string]*big.Int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDeriver(nil)
			require.NoError(t, err)

			// Make a copy of validators to avoid modifying the test data
			validatorsCopy := make(entity.Validators, len(tt.validators))
			for i, v := range tt.validators {
				validatorsCopy[i] = entity.Validator{
					Operator:    v.Operator,
					VotingPower: entity.ToVotingPower(new(big.Int).Set(v.VotingPower.Int)),
					IsActive:    v.IsActive,
					Keys:        append([]entity.ValidatorKey{}, v.Keys...),
					Vaults:      append([]entity.ValidatorVault{}, v.Vaults...),
				}
			}

			result := d.fillValidatorsActive(tt.config, validatorsCopy)

			// Check total voting power
			require.Equal(t, tt.expectedTotalVotingPower, result, "total voting power mismatch")

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

func TestDeriver_GetValidatorSet_HeaderCommitted(t *testing.T) {
	setup := setupBasicTest(t)
	setup.setupCommittedScenario()

	result, err := setup.deriver.GetValidatorSet(context.Background(), setup.epoch, setup.config)

	assertBasicValidatorSetProperties(t, result, err, setup.epoch, setup.timestamp)
	require.Equal(t, entity.HeaderCommitted, result.Status)
	require.Equal(t, common.HexToHash("0x111"), result.PreviousHeaderHash)
}

func TestDeriver_GetValidatorSet_HeaderPending(t *testing.T) {
	setup := setupBasicTest(t)
	setup.setupPendingScenario()

	result, err := setup.deriver.GetValidatorSet(context.Background(), setup.epoch, setup.config)

	assertBasicValidatorSetProperties(t, result, err, setup.epoch, setup.timestamp)
	require.Equal(t, entity.HeaderPending, result.Status)
	require.Equal(t, common.HexToHash("0x888"), result.PreviousHeaderHash)
}

// Helper functions to create test data
func createTestNetworkConfig() entity.NetworkConfig {
	return entity.NetworkConfig{
		VotingPowerProviders: []entity.CrossChainAddress{
			{ChainId: 1, Address: common.HexToAddress("0x123")},
		},
		KeysProvider: entity.CrossChainAddress{
			ChainId: 1, Address: common.HexToAddress("0x456"),
		},
		Replicas: []entity.CrossChainAddress{
			{ChainId: 1, Address: common.HexToAddress("0x789")},
		},
		RequiredHeaderKeyTag:    entity.KeyTag(15),
		MinInclusionVotingPower: entity.ToVotingPower(big.NewInt(100)),
		MaxVotingPower:          entity.ToVotingPower(big.NewInt(0)),
		MaxValidatorsCount:      entity.ToVotingPower(big.NewInt(0)),
		QuorumThresholds: []entity.QuorumThreshold{
			{
				KeyTag:          entity.KeyTag(15),
				QuorumThreshold: entity.ToQuorumThresholdPct(big.NewInt(670000000000000000)), // 67%
			},
		},
	}
}

func createTestOperatorVotingPowers() []entity.OperatorVotingPower {
	return []entity.OperatorVotingPower{
		{
			Operator: common.HexToAddress("0xabc"),
			Vaults: []entity.VaultVotingPower{
				{
					Vault:       common.HexToAddress("0xdef"),
					VotingPower: entity.ToVotingPower(big.NewInt(1000)),
				},
			},
		},
	}
}

func createTestOperatorWithKeys() []entity.OperatorWithKeys {
	return []entity.OperatorWithKeys{
		{
			Operator: common.HexToAddress("0xabc"),
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: entity.CompactPublicKey("testkey"),
				},
			},
		},
	}
}

type testSetup struct {
	mockClient *mocks.MockethClient
	deriver    *Deriver
	config     entity.NetworkConfig
	epoch      uint64
	timestamp  uint64
}

func setupBasicTest(t *testing.T) *testSetup {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := mocks.NewMockethClient(ctrl)
	config := createTestNetworkConfig()
	epoch := uint64(10)
	timestamp := uint64(1640995200)

	// Setup common mocks
	mockClient.EXPECT().GetEpochStart(gomock.Any(), epoch).Return(timestamp, nil)
	mockClient.EXPECT().GetVotingPowers(gomock.Any(), config.VotingPowerProviders[0], timestamp).Return(createTestOperatorVotingPowers(), nil)
	mockClient.EXPECT().GetKeys(gomock.Any(), config.KeysProvider, timestamp).Return(createTestOperatorWithKeys(), nil)

	deriver, err := NewDeriver(mockClient)
	require.NoError(t, err)

	return &testSetup{
		mockClient: mockClient,
		deriver:    deriver,
		config:     config,
		epoch:      epoch,
		timestamp:  timestamp,
	}
}

func (ts *testSetup) setupCommittedScenario() {
	ts.mockClient.EXPECT().IsValsetHeaderCommittedAt(gomock.Any(), ts.config.Replicas[0], ts.epoch).Return(true, nil)
	ts.mockClient.EXPECT().GetPreviousHeaderHashAt(gomock.Any(), ts.config.Replicas[0], ts.epoch).Return(common.HexToHash("0x111"), nil)
	ts.mockClient.EXPECT().GetHeaderHashAt(gomock.Any(), ts.config.Replicas[0], ts.epoch).Return(common.HexToHash("0xbf4eeff1b57d53e7d546e8339e7bac531abb6d22b147605fefeeb76886b43c9d"), nil)
}

func (ts *testSetup) setupPendingScenario() {
	ts.mockClient.EXPECT().IsValsetHeaderCommittedAt(gomock.Any(), ts.config.Replicas[0], ts.epoch).Return(false, nil)
	ts.mockClient.EXPECT().GetLastCommittedHeaderEpoch(gomock.Any(), ts.config.Replicas[0]).Return(uint64(8), nil)
	ts.mockClient.EXPECT().GetHeaderHash(gomock.Any(), ts.config.Replicas[0]).Return(common.HexToHash("0x888"), nil)
}

func assertBasicValidatorSetProperties(t *testing.T, result entity.ValidatorSet, err error, epoch uint64, timestamp uint64) {
	t.Helper()
	require.NoError(t, err)
	require.Equal(t, epoch, result.Epoch)
	require.Equal(t, timestamp, result.CaptureTimestamp)
	require.Len(t, result.Validators, 1)
	require.Equal(t, common.HexToAddress("0xabc"), result.Validators[0].Operator)
	require.True(t, result.Validators[0].IsActive)
}

func TestDeriver_GetNetworkData(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(client *mocks.MockethClient)
		addr       entity.CrossChainAddress
		expected   entity.NetworkData
		errorMsg   string
	}{
		{
			name: "successful get network data",
			setupMocks: func(m *mocks.MockethClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.HexToHash("0x456"), nil)
				m.EXPECT().GetEip712Domain(gomock.Any(), gomock.Any()).Return(entity.Eip712Domain{
					Name:    "TestNetwork",
					Version: "1",
				}, nil)
			},
			addr: entity.CrossChainAddress{},
			expected: entity.NetworkData{
				Address:    common.HexToAddress("0x123"),
				Subnetwork: common.HexToHash("0x456"),
				Eip712Data: entity.Eip712Domain{
					Name:    "TestNetwork",
					Version: "1",
				},
			},
		},
		{
			name: "network address error",
			setupMocks: func(m *mocks.MockethClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.Address{}, errors.New("network address error"))
			},
			addr:     entity.CrossChainAddress{},
			errorMsg: "failed to get network address",
		},
		{
			name: "subnetwork error",
			setupMocks: func(m *mocks.MockethClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.Hash{}, errors.New("subnetwork error"))
			},
			addr:     entity.CrossChainAddress{},
			errorMsg: "failed to get subnetwork",
		},
		{
			name: "eip712 domain error",
			setupMocks: func(m *mocks.MockethClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.HexToHash("0x456"), nil)
				m.EXPECT().GetEip712Domain(gomock.Any(), gomock.Any()).Return(entity.Eip712Domain{}, errors.New("eip712 error"))
			},
			addr:     entity.CrossChainAddress{},
			errorMsg: "failed to get eip712 domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockethClient(ctrl)
			tt.setupMocks(mockClient)

			d, err := NewDeriver(mockClient)
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
