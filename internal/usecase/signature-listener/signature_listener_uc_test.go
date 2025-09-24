package signature_listener

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	entity_processor "github.com/symbioticfi/relay/core/usecase/entity-processor/entity-processor"
	"github.com/symbioticfi/relay/core/usecase/entity-processor/entity-processor/mocks"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
	intEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/signals"
)

func TestHandleSignatureReceivedMessage_HappyPath(t *testing.T) {
	setup := newTestSetup(t)

	// Create real private key for signing
	privateKey := newPrivateKey(t)
	msg := "test-message-to-sign"

	// Create signature with the private key
	signature, hash, err := privateKey.Sign([]byte(msg))
	require.NoError(t, err)

	// Create validator set with the matching public key
	validatorSet := setup.createTestValidatorSetWithKey(t, privateKey)

	// Create P2P message with real signature
	p2pMsg := createTestP2PMessageWithSignature(privateKey, hash, signature)

	// Execute
	require.NoError(t, setup.useCase.HandleSignatureReceivedMessage(t.Context(), p2pMsg))

	// Verify that signature was saved
	signatures, err := setup.repo.GetAllSignatures(t.Context(), p2pMsg.Message.SignatureTargetID())
	require.NoError(t, err)
	require.Len(t, signatures, 1)

	// Verify the signature matches what we expect
	require.Equal(t, hash, signatures[0].MessageHash)
	require.Equal(t, signature, signatures[0].Signature)
	require.Equal(t, privateKey.PublicKey().Raw(), signatures[0].PublicKey)

	// Verify that signature map was updated
	signatureMap, err := setup.repo.GetSignatureMap(t.Context(), p2pMsg.Message.SignatureTargetID())
	require.NoError(t, err)
	require.Equal(t, 0, signatureMap.CurrentVotingPower.Cmp(validatorSet.Validators[0].VotingPower.Int))
}

type testSetup struct {
	repo    *badger.Repository
	useCase *SignatureListenerUseCase
}

func newTestSetup(t *testing.T) *testSetup {
	t.Helper()

	repo, err := badger.New(badger.Config{
		Dir: t.TempDir(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.Close()
	})

	// Create mock aggregator for entity processor
	ctrl := gomock.NewController(t)
	mockAggregator := mocks.NewMockAggregator(ctrl)
	mockAggregator.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	// Create mock aggregation proof signal for entity processor
	mockAggProofSignal := mocks.NewMockAggProofSignal(ctrl)
	mockAggProofSignal.EXPECT().Emit(gomock.Any()).Return(nil).AnyTimes()

	// Create mock signature processed signal for entity processor
	signatureProcessedSignal := signals.New[entity.SignatureExtended](signals.DefaultConfig(), "test", nil)

	processor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               mockAggregator,
		AggProofSignal:           mockAggProofSignal,
		SignatureProcessedSignal: signatureProcessedSignal,
	})
	require.NoError(t, err)

	cfg := Config{
		Repo:            repo,
		EntityProcessor: processor,
		SignalCfg: signals.Config{
			BufferSize:  10,
			WorkerCount: 5,
		},
		SelfP2PID: "test-self-p2p-id",
	}

	useCase, err := New(cfg)
	require.NoError(t, err)

	return &testSetup{
		repo:    repo,
		useCase: useCase,
	}
}

func newPrivateKey(t *testing.T) crypto.PrivateKey {
	t.Helper()
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	require.NoError(t, err)

	privateKey, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, privateKeyBytes)
	require.NoError(t, err)
	return privateKey
}

func (setup *testSetup) createTestValidatorSetWithKey(t *testing.T, privateKey crypto.PrivateKey) entity.ValidatorSet {
	t.Helper()
	vs := entity.ValidatorSet{
		Version:         1,
		RequiredKeyTag:  entity.KeyTag(15),
		Epoch:           1,
		QuorumThreshold: entity.ToVotingPower(big.NewInt(670)),
		Validators: []entity.Validator{{
			Operator:    common.HexToAddress("0x123"),
			VotingPower: entity.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys: []entity.ValidatorKey{
				{
					Tag:     entity.KeyTag(15),
					Payload: privateKey.PublicKey().OnChain(),
				},
			},
		}},
	}

	// Save the validator set to the repository
	err := setup.repo.SaveValidatorSet(t.Context(), vs)
	require.NoError(t, err)

	return vs
}

func createTestP2PMessageWithSignature(privateKey crypto.PrivateKey, hash []byte, signature []byte) intEntity.P2PMessage[entity.SignatureExtended] {
	return intEntity.P2PMessage[entity.SignatureExtended]{
		SenderInfo: intEntity.SenderInfo{
			Sender:    "test-peer-id",
			PublicKey: []byte("test-sender-pubkey"),
		},
		Message: entity.SignatureExtended{
			KeyTag:      entity.KeyTag(15),
			Epoch:       1,
			MessageHash: hash,
			PublicKey:   privateKey.PublicKey().Raw(),
			Signature:   signature,
		},
	}
}
