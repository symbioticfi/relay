package evm

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/client/evm/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	cryptoSym "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestOperatorTransactionsUseTargetChain(t *testing.T) {
	for name, call := range map[string]func(*Client, context.Context, symbiotic.CrossChainAddress) (symbiotic.TxResult, error){
		"register operator":       (*Client).RegisterOperator,
		"invalidate signatures":   (*Client).InvalidateOldSignatures,
		"register voting power":   (*Client).RegisterOperatorVotingPowerProvider,
		"unregister voting power": (*Client).UnregisterOperatorVotingPowerProvider,
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			backend := mocks.NewMockconn(ctrl)
			keys := mocks.NewMockkeyProvider(ctrl)
			target := symbiotic.CrossChainAddress{ChainId: 10, Address: common.HexToAddress("0x1234")}
			driver := symbiotic.CrossChainAddress{ChainId: 1, Address: common.HexToAddress("0x5678")}
			privateKey, err := crypto.GenerateKey()
			require.NoError(t, err)
			key, err := cryptoSym.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, crypto.FromECDSA(privateKey))
			require.NoError(t, err)
			keys.EXPECT().GetPrivateKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, int(target.ChainId)).Return(key, nil)
			client := &Client{
				cfg:           Config{DriverAddress: driver, KeyProvider: keys, RequestTimeout: time.Second, FallbackGasPrices: map[uint64]uint64{10: 3_000_000_000}},
				conns:         map[uint64]clientWithInfo{10: {conn: backend}, 1: {conn: mocks.NewMockconn(ctrl)}},
				driverChainID: driver.ChainId,
			}
			// The driver's backend has no expectations: every step must use the target.
			backend.EXPECT().PendingCodeAt(gomock.Any(), target.Address).Return([]byte{1}, nil)
			backend.EXPECT().EstimateGas(gomock.Any(), gomock.Any()).Return(uint64(50_000), nil)
			backend.EXPECT().PendingNonceAt(gomock.Any(), crypto.PubkeyToAddress(privateKey.PublicKey)).Return(uint64(7), nil)
			var sent *types.Transaction
			backend.EXPECT().SendTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, tx *types.Transaction) error {
				require.Equal(t, big.NewInt(10), tx.ChainId())
				require.Equal(t, &target.Address, tx.To())
				sent = tx
				return nil
			})
			backend.EXPECT().TransactionReceipt(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, hash common.Hash) (*types.Receipt, error) {
				require.Equal(t, sent.Hash(), hash)
				return &types.Receipt{TxHash: hash, Status: types.ReceiptStatusSuccessful}, nil
			})
			result, err := call(client, t.Context(), target)
			require.NoError(t, err)
			require.Equal(t, sent.Hash(), result.TxHash)
		})
	}
}
