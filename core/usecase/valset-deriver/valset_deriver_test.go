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
