package api_server

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	"github.com/symbioticfi/relay/internal/usecase/broadcaster"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type mockSignaturesStream struct {
	ctx        context.Context
	sentItems  []*apiv1.ListenSignaturesResponse
	sendError  error
	sendCalled chan struct{}
}

func (m *mockSignaturesStream) Context() context.Context {
	return m.ctx
}

func (m *mockSignaturesStream) Send(msg *apiv1.ListenSignaturesResponse) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentItems = append(m.sentItems, msg)
	if m.sendCalled != nil {
		select {
		case m.sendCalled <- struct{}{}:
		default:
		}
	}
	return nil
}

func (m *mockSignaturesStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockSignaturesStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockSignaturesStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockSignaturesStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockSignaturesStream) SetTrailer(metadata.MD) {
}

func TestListenSignatures_OnlyHistoricalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	expectedSignatures := []symbiotic.Signature{
		{
			MessageHash: common.Hex2Bytes("abcd1234"),
			KeyTag:      15,
			Epoch:       5,
			Signature:   common.Hex2Bytes("sig1"),
			PublicKey:   priv.PublicKey(),
		},
		{
			MessageHash: common.Hex2Bytes("efgh5678"),
			KeyTag:      15,
			Epoch:       5,
			Signature:   common.Hex2Bytes("sig2"),
			PublicKey:   priv.PublicKey(),
		},
		{
			MessageHash: common.Hex2Bytes("ijkl9012"),
			KeyTag:      15,
			Epoch:       6,
			Signature:   common.Hex2Bytes("sig3"),
			PublicKey:   priv.PublicKey(),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockSignaturesStream{ctx: ctx}

	startEpoch := uint64(5)
	mockRepo.EXPECT().GetSignaturesStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedSignatures, nil)

	req := &apiv1.ListenSignaturesRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenSignatures(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, []byte(expectedSignatures[0].MessageHash), stream.sentItems[0].GetSignature().GetMessageHash())
	require.Equal(t, []byte(expectedSignatures[1].MessageHash), stream.sentItems[1].GetSignature().GetMessageHash())
	require.Equal(t, []byte(expectedSignatures[2].MessageHash), stream.sentItems[2].GetSignature().GetMessageHash())
}

func TestListenSignatures_OnlyBroadcast(t *testing.T) {
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockSignaturesStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenSignaturesRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenSignatures(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newSignature := symbiotic.Signature{
		MessageHash: common.Hex2Bytes("newHash"),
		KeyTag:      15,
		Epoch:       10,
		Signature:   common.Hex2Bytes("newSig"),
		PublicKey:   priv.PublicKey(),
	}

	signatureHub.Broadcast(newSignature)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 1)
	require.Equal(t, []byte(newSignature.MessageHash), stream.sentItems[0].GetSignature().GetMessageHash())
	require.Equal(t, []byte(newSignature.Signature), stream.sentItems[0].GetSignature().GetSignature())
}

func TestListenSignatures_HistoricalAndBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	historicalSignatures := []symbiotic.Signature{
		{
			MessageHash: common.Hex2Bytes("hist1"),
			KeyTag:      15,
			Epoch:       5,
			Signature:   common.Hex2Bytes("histSig1"),
			PublicKey:   priv.PublicKey(),
		},
		{
			MessageHash: common.Hex2Bytes("hist2"),
			KeyTag:      15,
			Epoch:       5,
			Signature:   common.Hex2Bytes("histSig2"),
			PublicKey:   priv.PublicKey(),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockSignaturesStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	startEpoch := uint64(5)
	mockRepo.EXPECT().GetSignaturesStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(historicalSignatures, nil)

	req := &apiv1.ListenSignaturesRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenSignatures(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newSignature := symbiotic.Signature{
		MessageHash: common.Hex2Bytes("newHash"),
		KeyTag:      15,
		Epoch:       10,
		Signature:   common.Hex2Bytes("newSig"),
		PublicKey:   priv.PublicKey(),
	}

	signatureHub.Broadcast(newSignature)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, []byte(historicalSignatures[0].MessageHash), stream.sentItems[0].GetSignature().GetMessageHash())
	require.Equal(t, []byte(historicalSignatures[1].MessageHash), stream.sentItems[1].GetSignature().GetMessageHash())
	require.Equal(t, []byte(newSignature.MessageHash), stream.sentItems[2].GetSignature().GetMessageHash())
}

func TestListenSignatures_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockSignaturesStream{ctx: ctx}

	startEpoch := uint64(5)
	expectedError := errors.New("database connection failed")
	mockRepo.EXPECT().GetSignaturesStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(nil, expectedError)

	req := &apiv1.ListenSignaturesRequest{
		StartEpoch: &startEpoch,
	}

	err := handler.ListenSignatures(req, stream)

	require.Error(t, err)
	require.Equal(t, expectedError, err)
	require.Empty(t, stream.sentItems)
}

func TestListenSignatures_StreamSendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	expectedSignatures := []symbiotic.Signature{
		{
			MessageHash: common.Hex2Bytes("abcd1234"),
			KeyTag:      15,
			Epoch:       5,
			Signature:   common.Hex2Bytes("sig1"),
			PublicKey:   priv.PublicKey(),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sendError := errors.New("stream send failed")
	stream := &mockSignaturesStream{
		ctx:       ctx,
		sendError: sendError,
	}

	startEpoch := uint64(5)
	mockRepo.EXPECT().GetSignaturesStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedSignatures, nil)

	req := &apiv1.ListenSignaturesRequest{
		StartEpoch: &startEpoch,
	}

	err = handler.ListenSignatures(req, stream)

	require.Error(t, err)
	require.Equal(t, sendError, err)
}

func TestListenSignatures_MultipleBroadcasts(t *testing.T) {
	signatureHub := broadcaster.NewHub[symbiotic.Signature]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		signatureHub: signatureHub,
	}

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockSignaturesStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenSignaturesRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenSignatures(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		newSignature := symbiotic.Signature{
			MessageHash: common.Hex2Bytes(string(rune(i))),
			KeyTag:      15,
			Epoch:       symbiotic.Epoch(10 + i),
			Signature:   common.Hex2Bytes(string(rune(i))),
			PublicKey:   priv.PublicKey(),
		}
		signatureHub.Broadcast(newSignature)
		<-stream.sendCalled
	}

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 5)
}
