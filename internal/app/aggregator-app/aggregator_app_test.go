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
	"middleware-offchain/pkg/bls"
)

func TestHandleSignatureGeneratedMessage(t *testing.T) {
	t.Skip("need fix")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyTag := entity.KeyTag(7)
	operatorAddress1 := common.BytesToAddress([]byte{1})
	operatorAddress2 := common.BytesToAddress([]byte{2})
	operatorAddress3 := common.BytesToAddress([]byte{3})

	bi1, ok := big.NewInt(0).SetString("87191036493798670866484781455694320176667203290824056510541300741498740913410", 10)
	require.True(t, ok)
	bi2, ok := big.NewInt(0).SetString("11008377096554045051122023680185802911050337017631086444859313200352654461863", 10)
	require.True(t, ok)
	bi3, ok := big.NewInt(0).SetString("26972876870930381973856869753776124637336739336929668162870464864826929175089", 10)
	require.True(t, ok)

	key1 := bls.ComputeKeyPair(bi1.Bytes())
	key2 := bls.ComputeKeyPair(bi2.Bytes())
	key3 := bls.ComputeKeyPair(bi3.Bytes())

	bytes := [32]byte{1, 2, 3}
	sign1, err := key1.Sign(bytes[:])
	require.NoError(t, err)
	sign2, err := key2.Sign(bytes[:])
	require.NoError(t, err)
	sign3, err := key3.Sign(bytes[:])
	require.NoError(t, err)

	mockEthClient := mocks.NewMockethClient(ctrl)
	mockValsetDeriver := mocks.NewMockvalsetDeriver(ctrl)
	mockP2P := mocks.NewMockp2pClient(ctrl)
	mockP2P.EXPECT().SetSignatureHashMessageHandler(gomock.Any())

	validatorSet := entity.ValidatorSet{
		Validators: []entity.Validator{
			{
				Operator:    operatorAddress1,
				IsActive:    true,
				VotingPower: big.NewInt(101),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key1.PublicKeyG1.Marshal(),
				}},
			},
			{
				Operator:    operatorAddress2,
				IsActive:    true,
				VotingPower: big.NewInt(201),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key2.PublicKeyG1.Marshal(),
				}},
			},
			{
				Operator:    operatorAddress3,
				IsActive:    true,
				VotingPower: big.NewInt(301),
				Keys: []entity.Key{{
					Tag:     keyTag,
					Payload: key3.PublicKeyG1.Marshal(),
				}},
			},
		},
	}
	mockValsetDeriver.EXPECT().
		GetValidatorSet(gomock.Any(), gomock.Any()).
		Return(validatorSet, nil)

	cfg := Config{
		P2PClient: mockP2P,
	}

	ctx := t.Context()

	app, err := NewAggregatorApp(cfg)
	require.NoError(t, err)

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
					PublicKey: []byte{0x02},
					//KeyTag:    keyTag,
				},
			},
			mockSetup:   func() {},
			expectedErr: errors.New("validator not found for public key: 02"),
		},
		{
			name: "quorum not reached",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKey: key1.PackPublicG1G2(),
					//KeyTag:    keyTag,
					Signature: bls.SerializeG1(sign1),
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
					PublicKey: key3.PackPublicG1G2(),
					//KeyTag:    keyTag,
					Signature: bls.SerializeG1(sign3),
				},
			},
			mockSetup: func() {
				mockEthClient.EXPECT().
					GetQuorumThreshold(ctx, gomock.Any(), keyTag).
					Return(big.NewInt(500000000000000000), nil)

				mockP2P.EXPECT().BroadcastSignatureAggregatedMessage(ctx, gomock.Any()).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "error getting quorum threshold",
			message: entity.P2PSignatureHashMessage{
				Message: entity.SignatureHashMessage{
					PublicKey: key2.PackPublicG1G2(),
					//KeyTag:    keyTag,
					Signature: bls.SerializeG1(sign2),
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
