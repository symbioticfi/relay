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
)

type mockProofsStream struct {
	ctx        context.Context
	sentItems  []*apiv1.ListenProofsResponse
	sendError  error
	sendCalled chan struct{}
}

func (m *mockProofsStream) Context() context.Context {
	return m.ctx
}

func (m *mockProofsStream) Send(msg *apiv1.ListenProofsResponse) error {
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

func (m *mockProofsStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockProofsStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockProofsStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockProofsStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockProofsStream) SetTrailer(metadata.MD) {
}

func TestListenProofs_OnlyHistoricalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	expectedProofs := []symbiotic.AggregationProof{
		{
			MessageHash: common.Hex2Bytes("hash1"),
			KeyTag:      15,
			Epoch:       3,
			Proof:       common.Hex2Bytes("proof1"),
		},
		{
			MessageHash: common.Hex2Bytes("hash2"),
			KeyTag:      15,
			Epoch:       3,
			Proof:       common.Hex2Bytes("proof2"),
		},
		{
			MessageHash: common.Hex2Bytes("hash3"),
			KeyTag:      15,
			Epoch:       4,
			Proof:       common.Hex2Bytes("proof3"),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{ctx: ctx}

	startEpoch := uint64(3)
	mockRepo.EXPECT().GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedProofs, nil)

	req := &apiv1.ListenProofsRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenProofs(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, []byte(expectedProofs[0].MessageHash), stream.sentItems[0].GetAggregationProof().GetMessageHash())
	require.Equal(t, []byte(expectedProofs[1].MessageHash), stream.sentItems[1].GetAggregationProof().GetMessageHash())
	require.Equal(t, []byte(expectedProofs[2].MessageHash), stream.sentItems[2].GetAggregationProof().GetMessageHash())
}

func TestListenProofs_OnlyBroadcast(t *testing.T) {
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenProofsRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenProofs(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newProof := symbiotic.AggregationProof{
		MessageHash: common.Hex2Bytes("newHash"),
		KeyTag:      15,
		Epoch:       10,
		Proof:       common.Hex2Bytes("newProof"),
	}

	proofsHub.Broadcast(newProof)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 1)
	require.Equal(t, []byte(newProof.MessageHash), stream.sentItems[0].GetAggregationProof().GetMessageHash())
	require.Equal(t, []byte(newProof.Proof), stream.sentItems[0].GetAggregationProof().GetProof())
}

func TestListenProofs_HistoricalAndBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	historicalProofs := []symbiotic.AggregationProof{
		{
			MessageHash: common.Hex2Bytes("hist1"),
			KeyTag:      15,
			Epoch:       3,
			Proof:       common.Hex2Bytes("histProof1"),
		},
		{
			MessageHash: common.Hex2Bytes("hist2"),
			KeyTag:      15,
			Epoch:       3,
			Proof:       common.Hex2Bytes("histProof2"),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	startEpoch := uint64(3)
	mockRepo.EXPECT().GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(historicalProofs, nil)

	req := &apiv1.ListenProofsRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenProofs(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newProof := symbiotic.AggregationProof{
		MessageHash: common.Hex2Bytes("newHash"),
		KeyTag:      15,
		Epoch:       10,
		Proof:       common.Hex2Bytes("newProof"),
	}

	proofsHub.Broadcast(newProof)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, []byte(historicalProofs[0].MessageHash), stream.sentItems[0].GetAggregationProof().GetMessageHash())
	require.Equal(t, []byte(historicalProofs[1].MessageHash), stream.sentItems[1].GetAggregationProof().GetMessageHash())
	require.Equal(t, []byte(newProof.MessageHash), stream.sentItems[2].GetAggregationProof().GetMessageHash())
}

func TestListenProofs_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{ctx: ctx}

	startEpoch := uint64(3)
	expectedError := errors.New("database connection failed")
	mockRepo.EXPECT().GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(nil, expectedError)

	req := &apiv1.ListenProofsRequest{
		StartEpoch: &startEpoch,
	}

	err := handler.ListenProofs(req, stream)

	require.Error(t, err)
	require.Equal(t, expectedError, err)
	require.Empty(t, stream.sentItems)
}

func TestListenProofs_StreamSendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	expectedProofs := []symbiotic.AggregationProof{
		{
			MessageHash: common.Hex2Bytes("hash1"),
			KeyTag:      15,
			Epoch:       3,
			Proof:       common.Hex2Bytes("proof1"),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sendError := errors.New("stream send failed")
	stream := &mockProofsStream{
		ctx:       ctx,
		sendError: sendError,
	}

	startEpoch := uint64(3)
	mockRepo.EXPECT().GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedProofs, nil)

	req := &apiv1.ListenProofsRequest{
		StartEpoch: &startEpoch,
	}

	err := handler.ListenProofs(req, stream)

	require.Error(t, err)
	require.Equal(t, sendError, err)
}

func TestListenProofs_MultipleBroadcasts(t *testing.T) {
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenProofsRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenProofs(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		newProof := symbiotic.AggregationProof{
			MessageHash: common.Hex2Bytes(string(rune(i))),
			KeyTag:      15,
			Epoch:       symbiotic.Epoch(10 + i),
			Proof:       common.Hex2Bytes(string(rune(i))),
		}
		proofsHub.Broadcast(newProof)
		<-stream.sendCalled
	}

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 5)
}

func TestListenProofs_EmptyHistoricalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	proofsHub := broadcaster.NewHub[symbiotic.AggregationProof]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		proofsHub: proofsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockProofsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	startEpoch := uint64(3)
	mockRepo.EXPECT().GetAggregationProofsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return([]symbiotic.AggregationProof{}, nil)

	req := &apiv1.ListenProofsRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenProofs(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newProof := symbiotic.AggregationProof{
		MessageHash: common.Hex2Bytes("newHash"),
		KeyTag:      15,
		Epoch:       10,
		Proof:       common.Hex2Bytes("newProof"),
	}

	proofsHub.Broadcast(newProof)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 1)
	require.Equal(t, []byte(newProof.MessageHash), stream.sentItems[0].GetAggregationProof().GetMessageHash())
}
