package valset

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/mocks"
)

func TestValsetDeriver_GetValidatorSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEthClient := mocks.NewMockethClient(ctrl)

	valsetDeriver, err := NewValsetDeriver(mockEthClient)
	require.NoError(t, err)

	ctx := t.Context()

	vpAddress := common.BytesToAddress([]byte{1})
	kpAddress := common.BytesToAddress([]byte{2})
	operatorAddress := common.BytesToAddress([]byte{3})
	vaultAddress := common.BytesToAddress([]byte{4})

	timestamp := big.NewInt(1234567890)
	keyPayload := []byte{1, 2, 3, 4}
	keyTag := uint8(6)

	tests := []struct {
		name        string
		mockSetup   func()
		expectedSet entity.ValidatorSet
		expectedErr error
	}{
		{
			name: "successfully fetch validator set",
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetMasterConfig(ctx, timestamp).
					Return(entity.MasterConfig{
						VotingPowerProviders: []entity.CrossChainAddress{{Address: vpAddress}},
						KeysProvider:         entity.CrossChainAddress{Address: kpAddress},
					}, nil)

				mockEthClient.EXPECT().
					GetValSetConfig(ctx, timestamp).
					Return(entity.ValSetConfig{
						MinInclusionVotingPower: big.NewInt(100),
						MaxVotingPower:          big.NewInt(1000),
						MaxValidatorsCount:      big.NewInt(10),
					}, nil)

				mockEthClient.EXPECT().
					GetVotingPowers(ctx, vpAddress, timestamp).
					Return([]entity.OperatorVotingPower{{
						Operator: operatorAddress,
						Vaults: []entity.VaultVotingPower{{
							Vault:       vaultAddress,
							VotingPower: big.NewInt(500),
						}},
					}}, nil)

				mockEthClient.EXPECT().
					GetKeys(ctx, kpAddress, timestamp).
					Return([]entity.OperatorWithKeys{{
						Operator: operatorAddress,
						Keys: []entity.Key{
							{
								Tag:     keyTag,
								Payload: keyPayload,
							},
						},
					}}, nil)
			},
			expectedSet: entity.ValidatorSet{
				Version: 1,
				Validators: []entity.Validator{{
					Operator:    operatorAddress,
					VotingPower: big.NewInt(500),
					IsActive:    true,
					Keys: []entity.Key{{
						Tag:     keyTag,
						Payload: keyPayload,
					}},
					Vaults: []entity.Vault{{
						Vault:       vaultAddress,
						VotingPower: big.NewInt(500),
					}},
				}},
				TotalActiveVotingPower: big.NewInt(500),
			},
			expectedErr: nil,
		},
		{
			name: "error fetching master config",
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetMasterConfig(ctx, timestamp).
					Return(entity.MasterConfig{}, errors.New("failed to fetch master config"))
			},
			expectedSet: entity.ValidatorSet{},
			expectedErr: errors.New("failed to get master config: failed to fetch master config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := valsetDeriver.GetValidatorSet(ctx, timestamp)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSet, result)
			}
		})
	}
}
