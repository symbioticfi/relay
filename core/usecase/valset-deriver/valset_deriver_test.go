package valsetDeriver

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/core/usecase/ssz"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/valset-deriver/mocks"
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
		addr       entity.CrossChainAddress
		expected   entity.NetworkData
		errorMsg   string
	}{
		{
			name: "successful get network data",
			setupMocks: func(m *mocks.MockEvmClient) {
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
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.Address{}, errors.New("network address error"))
			},
			addr:     entity.CrossChainAddress{},
			errorMsg: "failed to get network address",
		},
		{
			name: "subnetwork error",
			setupMocks: func(m *mocks.MockEvmClient) {
				m.EXPECT().GetNetworkAddress(gomock.Any()).Return(common.HexToAddress("0x123"), nil)
				m.EXPECT().GetSubnetwork(gomock.Any()).Return(common.Hash{}, errors.New("subnetwork error"))
			},
			addr:     entity.CrossChainAddress{},
			errorMsg: "failed to get subnetwork",
		},
		{
			name: "eip712 domain error",
			setupMocks: func(m *mocks.MockEvmClient) {
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

			mockEvmClient := mocks.NewMockEvmClient(ctrl)
			tt.setupMocks(mockEvmClient)

			d, err := NewDeriver(mockEvmClient)
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
	// This test covers the scenario where a validator has more vaults than ssz.VaultsListMaxElements
	// The function should only keep the top vaults (sorted by voting power desc, then operator address asc)

	// Create voting powers with more vaults than the SSZ limit
	// Assuming ssz.VaultsListMaxElements is a reasonable number like 256 or 512
	// We'll create more vaults than this limit to test the truncation logic

	numVaults := ssz.VaultsListMaxElements + 10 // Create more than the limit
	vaultVotingPowers := make([]entity.VaultVotingPower, numVaults)

	// Create vaults with decreasing voting power to test sorting
	for i := 0; i < numVaults; i++ {
		vaultVotingPowers[i] = entity.VaultVotingPower{
			Vault:       common.HexToAddress(fmt.Sprintf("0x%040d", i+1)),       // 0x000...001, 0x000...002, etc.
			VotingPower: entity.ToVotingPower(big.NewInt(int64(numVaults - i))), // Decreasing power: numVaults, numVaults-1, ..., 1
		}
	}

	votingPowers := []dtoOperatorVotingPower{
		{
			chainId: 1,
			votingPowers: []entity.OperatorVotingPower{
				{
					Operator: common.HexToAddress("0x123"),
					Vaults:   vaultVotingPowers,
				},
			},
		},
	}

	keys := []entity.OperatorWithKeys{
		{
			Operator: common.HexToAddress("0x123"),
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: entity.CompactPublicKey("key1"),
				},
			},
		},
	}

	result := fillValidators(votingPowers, keys)

	require.Len(t, result, 1)
	validator := result[0]

	// Verify that the number of vaults is limited to ssz.VaultsListMaxElements
	require.Len(t, validator.Vaults, ssz.VaultsListMaxElements)

	// Verify that the total voting power is calculated correctly
	// It should be the sum of all original vaults (before truncation)
	expectedTotalVotingPower := big.NewInt(0)
	for i := 1; i <= numVaults; i++ {
		expectedTotalVotingPower.Add(expectedTotalVotingPower, big.NewInt(int64(i)))
	}
	require.Equal(t, expectedTotalVotingPower, validator.VotingPower.Int)

	// Verify that the kept vaults are the ones with highest voting power
	// Since we created vaults with decreasing power, the first ssz.VaultsListMaxElements should be kept
	for i, vault := range validator.Vaults {
		expectedVotingPower := big.NewInt(int64(numVaults - i))
		require.Equal(t, expectedVotingPower, vault.VotingPower.Int,
			"Vault at index %d should have voting power %d", i, numVaults-i)
		require.Equal(t, uint64(1), vault.ChainID)
	}

	// Verify that vaults are sorted correctly (by voting power desc, then operator address asc)
	for i := 1; i < len(validator.Vaults); i++ {
		// Since all vaults have different voting powers in our test,
		// they should be sorted by voting power descending
		require.True(t, validator.Vaults[i-1].VotingPower.Int.Cmp(validator.Vaults[i].VotingPower.Int) > 0,
			"Vaults should be sorted by voting power descending")
	}

	// Verify other validator properties
	require.Equal(t, common.HexToAddress("0x123"), validator.Operator)
	require.False(t, validator.IsActive) // Should be false by default
	require.Len(t, validator.Keys, 1)
	require.Equal(t, entity.KeyTag(15), validator.Keys[0].Tag)
}
