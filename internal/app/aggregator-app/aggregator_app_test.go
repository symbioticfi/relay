package aggregator_app

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"middleware-offchain/internal/app/aggregator-app/mocks"
	"middleware-offchain/internal/entity"
)

func TestHandleSignatureGeneratedMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyTag := uint8(7)
	operatorAddress1 := common.BytesToAddress([]byte{1})
	operatorAddress2 := common.BytesToAddress([]byte{2})
	operatorAddress3 := common.BytesToAddress([]byte{3})
	key1 := []byte{1, 2, 3}
	key2 := []byte{3, 2, 1}
	key3 := []byte{7, 7, 7}

	mockEthClient := mocks.NewMockethClient(ctrl)
	mockValsetDeriver := mocks.NewMockvalsetDeriver(ctrl)

	validatorSet := entity.ValidatorSet{
		Validators: []entity.Validator{
			{
				Operator:    operatorAddress1,
				IsActive:    true,
				VotingPower: big.NewInt(101),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key1,
				}},
			},
			{
				Operator:    operatorAddress2,
				IsActive:    true,
				VotingPower: big.NewInt(201),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key2,
				}},
			},
			{
				Operator:    operatorAddress3,
				IsActive:    true,
				VotingPower: big.NewInt(301),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key3,
				}},
			},
		},
		TotalActiveVotingPower: big.NewInt(101 + 201 + 301),
	}
	mockValsetDeriver.EXPECT().
		GetValidatorSet(gomock.Any(), gomock.Any()).
		Return(validatorSet, nil)

	cfg := Config{
		EthClient:     mockEthClient,
		ValsetDeriver: mockValsetDeriver,
	}

	ctx := t.Context()

	app, err := NewAggregatorApp(ctx, cfg)
	require.NoError(t, err)

	app.validatorSet = validatorSet

	tests := []struct {
		name           string
		message        entity.P2PSignatureHashMessage
		mockSetup      func()
		expectedErr    error
		expectedLogMsg string
	}{
		{
			name: "validator not found",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKeyG1: []byte{0x02},
					KeyTag:      keyTag,
				},
			},
			mockSetup:   func() {},
			expectedErr: errors.New("validator not found for public key: 02"),
		},
		{
			name: "quorum not reached",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKeyG1: key1,
					KeyTag:      keyTag,
				},
			},
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetQuorumThreshold(ctx, gomock.Any(), keyTag).
					Return(big.NewInt(2000000000000000000), nil)
			},
			expectedErr: nil,
		},
		{
			name: "quorum reached",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKeyG1: key3,
					KeyTag:      keyTag,
				},
			},
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetQuorumThreshold(ctx, gomock.Any(), keyTag).
					Return(big.NewInt(500000000000000000), nil)
			},
			expectedErr: nil,
		},
		{
			name: "error getting quorum threshold",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKeyG1: key2,
					KeyTag:      keyTag,
				},
			},
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetQuorumThreshold(ctx, gomock.Any(), keyTag).
					Return(nil, errors.New("failed to get quorum threshold"))
			},
			expectedErr: errors.New("failed to get quorum threshold: failed to get quorum threshold"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := app.HandleSignatureGeneratedMessage(ctx, tt.message)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
