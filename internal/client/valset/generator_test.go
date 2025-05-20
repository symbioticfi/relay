package valset

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"middleware-offchain/internal/client/valset/mocks"
	"middleware-offchain/internal/entity"
)

func TestGenerator_GenerateValidatorSetHeaderHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeriver := mocks.NewMockderiver(ctrl)
	mockEthClient := mocks.NewMockethClient(ctrl)

	v := entity.ValidatorSetHeader{
		Version: 1,
		ActiveAggregatedKeys: []entity.Key{{
			Tag:     15,
			Payload: decodeHex(t, "264621561abeb4dac9a497cb21f305b8f41b56389734832656d7c7adde2247081ffa73b25b82c16096babd6a15d259a24a8304cd96ee6c27e790ff27d8744a5b"),
		}},
		TotalActiveVotingPower: new(big.Int).SetInt64(30000000000000),
		ValidatorsSszMRoot:     [32]byte(decodeHex(t, "d9354a3cf52fba5126422c86d35db53d566d46f9208faa86c7b9155d7dcf3926")),
		ExtraData:              decodeHex(t, "2695ed079545bb906f5868716071ab237e36d04fdc1aa07b06bd98c81185067d"),
	}

	eip := &entity.Eip712Domain{
		Name:    "Middleware",
		Version: "1",
		ChainId: new(big.Int).SetInt64(111),
	}

	mockEthClient.EXPECT().
		GetEip712Domain(t.Context()).
		Return(eip, nil)
	mockEthClient.EXPECT().
		GetCurrentEpoch(t.Context()).
		Return(new(big.Int).SetInt64(1), nil)
	mockEthClient.EXPECT().
		GetSubnetwork(t.Context()).
		Return(decodeHex(t, "f39fd6e51aad88f6f4ce6ab8827279cfffb92266000000000000000000000000"), nil)

	generator, err := NewGenerator(mockDeriver, mockEthClient)
	require.NoError(t, err)

	hashBytes, err := generator.GenerateValidatorSetHeaderHash(t.Context(), v)
	require.NoError(t, err)

	inContract := "a296e61b893375cafbc989aff8eef893b604237d12ea7a4a5912b99b4372e0eb"

	require.Equal(t, inContract, hex.EncodeToString(hashBytes))
}

func decodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}
