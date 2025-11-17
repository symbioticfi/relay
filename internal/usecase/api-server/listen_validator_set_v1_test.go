package api_server

import (
	"context"
	"testing"
	"time"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	"github.com/symbioticfi/relay/internal/usecase/broadcaster"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type mockValidatorSetsStream struct {
	ctx        context.Context
	sentItems  []*apiv1.ListenValidatorSetResponse
	sendError  error
	sendCalled chan struct{}
}

func (m *mockValidatorSetsStream) Context() context.Context {
	return m.ctx
}

func (m *mockValidatorSetsStream) Send(msg *apiv1.ListenValidatorSetResponse) error {
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

func (m *mockValidatorSetsStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockValidatorSetsStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockValidatorSetsStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockValidatorSetsStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockValidatorSetsStream) SetTrailer(metadata.MD) {
}

func TestListenValidatorSet_OnlyHistoricalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	expectedValidatorSets := []symbiotic.ValidatorSet{
		createTestValidatorSet(1),
		createTestValidatorSet(2),
		createTestValidatorSet(3),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{ctx: ctx}

	startEpoch := uint64(1)
	mockRepo.EXPECT().GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedValidatorSets, nil)

	req := &apiv1.ListenValidatorSetRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, uint64(expectedValidatorSets[0].Epoch), stream.sentItems[0].GetValidatorSet().GetEpoch())
	require.Equal(t, uint64(expectedValidatorSets[1].Epoch), stream.sentItems[1].GetValidatorSet().GetEpoch())
	require.Equal(t, uint64(expectedValidatorSets[2].Epoch), stream.sentItems[2].GetValidatorSet().GetEpoch())
}

func TestListenValidatorSet_OnlyBroadcast(t *testing.T) {
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenValidatorSetRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newValidatorSet := createTestValidatorSet(10)

	validatorSetsHub.Broadcast(newValidatorSet)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 1)
	require.Equal(t, uint64(newValidatorSet.Epoch), stream.sentItems[0].GetValidatorSet().GetEpoch())
}

func TestListenValidatorSet_HistoricalAndBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	historicalValidatorSets := []symbiotic.ValidatorSet{
		createTestValidatorSet(1),
		createTestValidatorSet(2),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	startEpoch := uint64(1)
	mockRepo.EXPECT().GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(historicalValidatorSets, nil)

	req := &apiv1.ListenValidatorSetRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newValidatorSet := createTestValidatorSet(10)

	validatorSetsHub.Broadcast(newValidatorSet)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 3)
	require.Equal(t, uint64(historicalValidatorSets[0].Epoch), stream.sentItems[0].GetValidatorSet().GetEpoch())
	require.Equal(t, uint64(historicalValidatorSets[1].Epoch), stream.sentItems[1].GetValidatorSet().GetEpoch())
	require.Equal(t, uint64(newValidatorSet.Epoch), stream.sentItems[2].GetValidatorSet().GetEpoch())
}

func TestListenValidatorSet_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{ctx: ctx}

	startEpoch := uint64(1)
	expectedError := errors.New("database connection failed")
	mockRepo.EXPECT().GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(nil, expectedError)

	req := &apiv1.ListenValidatorSetRequest{
		StartEpoch: &startEpoch,
	}

	err := handler.ListenValidatorSet(req, stream)

	require.Error(t, err)
	require.Equal(t, expectedError, err)
	require.Empty(t, stream.sentItems)
}

func TestListenValidatorSet_StreamSendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	expectedValidatorSets := []symbiotic.ValidatorSet{
		createTestValidatorSet(1),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sendError := errors.New("stream send failed")
	stream := &mockValidatorSetsStream{
		ctx:       ctx,
		sendError: sendError,
	}

	startEpoch := uint64(1)
	mockRepo.EXPECT().GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return(expectedValidatorSets, nil)

	req := &apiv1.ListenValidatorSetRequest{
		StartEpoch: &startEpoch,
	}

	err := handler.ListenValidatorSet(req, stream)

	require.Error(t, err)
	require.Equal(t, sendError, err)
}

func TestListenValidatorSet_MultipleBroadcasts(t *testing.T) {
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	req := &apiv1.ListenValidatorSetRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		newValidatorSet := createTestValidatorSet(symbiotic.Epoch(10 + i))
		validatorSetsHub.Broadcast(newValidatorSet)
		<-stream.sendCalled
	}

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 5)
	for i := 0; i < 5; i++ {
		require.Equal(t, uint64(10+i), stream.sentItems[i].GetValidatorSet().GetEpoch())
	}
}

func TestListenValidatorSet_EmptyHistoricalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			Repo:                   mockRepo,
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 10),
	}

	startEpoch := uint64(1)
	mockRepo.EXPECT().GetValidatorSetsStartingFromEpoch(ctx, symbiotic.Epoch(startEpoch)).Return([]symbiotic.ValidatorSet{}, nil)

	req := &apiv1.ListenValidatorSetRequest{
		StartEpoch: &startEpoch,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	newValidatorSet := createTestValidatorSet(10)

	validatorSetsHub.Broadcast(newValidatorSet)
	<-stream.sendCalled

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, 1)
	require.Equal(t, uint64(newValidatorSet.Epoch), stream.sentItems[0].GetValidatorSet().GetEpoch())
}

func TestListenValidatorSet_ConcurrentBroadcasts(t *testing.T) {
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 10,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := &mockValidatorSetsStream{
		ctx:        ctx,
		sendCalled: make(chan struct{}, 20),
	}

	req := &apiv1.ListenValidatorSetRequest{}

	errCh := make(chan error, 1)
	go func() {
		errCh <- handler.ListenValidatorSet(req, stream)
	}()

	time.Sleep(50 * time.Millisecond)

	broadcastCount := 10
	for i := 0; i < broadcastCount; i++ {
		go func(epoch int) {
			validatorSetsHub.Broadcast(createTestValidatorSet(symbiotic.Epoch(epoch)))
		}(i)
	}

	for i := 0; i < broadcastCount; i++ {
		<-stream.sendCalled
	}

	cancel()
	<-errCh

	require.Len(t, stream.sentItems, broadcastCount)
}

func TestListenValidatorSet_MaxStreamsReached_ReturnsError(t *testing.T) {
	validatorSetsHub := broadcaster.NewHub[symbiotic.ValidatorSet]()

	handler := &grpcHandler{
		cfg: Config{
			MaxAllowedStreamsCount: 0,
		},
		validatorSetsHub: validatorSetsHub,
	}

	ctx := context.Background()
	stream := &mockValidatorSetsStream{ctx: ctx}
	req := &apiv1.ListenValidatorSetRequest{}

	err := handler.ListenValidatorSet(req, stream)

	require.Error(t, err)
	require.Contains(t, err.Error(), "max allowed streams limit reached")
}
