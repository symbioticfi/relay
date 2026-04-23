package pruner

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/usecase/pruner/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestPruner_DripPruning_SingleEpoch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	ctx := context.Background()

	batchSize := 2
	requestIDs := []common.Hash{
		common.HexToHash("0x01"),
		common.HexToHash("0x02"),
		common.HexToHash("0x03"),
	}

	service := &Service{
		cfg: Config{
			Repo:                     mockRepo,
			Metrics:                  mockMetrics,
			SignatureRetentionEpochs: 4,
			PruneBatchSize:           batchSize,
		},
	}

	// Tick 1: returns 2 requestIDs (batch size), processes them
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(requestIDs[:2], nil)
	mockRepo.EXPECT().PruneSignaturesByRequestIDs(gomock.Any(), symbiotic.Epoch(0), requestIDs[:2], batchSize).Return(nil)

	done, err := service.pruneSignatureBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.False(t, done)

	// Tick 2: returns 1 requestID (remaining), processes it
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(requestIDs[2:], nil)
	mockRepo.EXPECT().PruneSignaturesByRequestIDs(gomock.Any(), symbiotic.Epoch(0), requestIDs[2:], batchSize).Return(nil)

	done, err = service.pruneSignatureBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.False(t, done)

	// Tick 3: returns 0 requestIDs, epoch is done
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(nil, nil)

	done, err = service.pruneSignatureBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.True(t, done)
}

func TestPruner_DripPruning_ProofBatch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	ctx := context.Background()

	batchSize := 10
	requestIDs := []common.Hash{common.HexToHash("0x01")}

	service := &Service{
		cfg: Config{
			Repo:                 mockRepo,
			Metrics:              mockMetrics,
			ProofRetentionEpochs: 4,
			PruneBatchSize:       batchSize,
		},
	}

	// Tick 1: has requestIDs, prune proofs
	mockRepo.EXPECT().PruneProofCommits(gomock.Any(), symbiotic.Epoch(0)).Return(nil)
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(requestIDs, nil)
	mockRepo.EXPECT().PruneProofsByRequestIDs(gomock.Any(), symbiotic.Epoch(0), requestIDs, batchSize).Return(nil)

	done, err := service.pruneProofBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.False(t, done)

	// Tick 2: no more requestIDs
	mockRepo.EXPECT().PruneProofCommits(gomock.Any(), symbiotic.Epoch(0)).Return(nil)
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(nil, nil)

	done, err = service.pruneProofBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.True(t, done)
}

func TestPruner_PruneEntityType_SkipsWhenInRetentionWindow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	service := &Service{
		cfg: Config{
			Repo:                     mockRepo,
			Metrics:                  mockMetrics,
			SignatureRetentionEpochs: 4,
			PruneBatchSize:           10,
		},
	}

	// latestEpoch=10, retention=4, oldestStored=7
	// oldestToKeep = 10 - 4 + 1 = 7, 7 >= 7 → nothing to prune
	service.pruneEntityType(
		context.Background(),
		symbiotic.Epoch(10), symbiotic.Epoch(7),
		4, "signature",
		func(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
			t.Fatal("should not be called")
			return false, nil
		},
	)
}

func TestPruner_PruneEntityType_DoneMovesToNextEpoch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	service := &Service{
		cfg: Config{
			Repo:                     mockRepo,
			Metrics:                  mockMetrics,
			SignatureRetentionEpochs: 4,
			PruneBatchSize:           10,
		},
	}

	// latestEpoch=10, retention=4, oldestStored=5
	// oldestToKeep = 10 - 4 + 1 = 7
	// epochs to prune: 5, 6
	var calledEpochs []symbiotic.Epoch
	mockMetrics.EXPECT().IncPrunedEpochsCount("test").Times(2)

	service.pruneEntityType(
		context.Background(),
		symbiotic.Epoch(10), symbiotic.Epoch(5),
		4, "test",
		func(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
			calledEpochs = append(calledEpochs, epoch)
			return true, nil // always done immediately
		},
	)

	require.Equal(t, []symbiotic.Epoch{5, 6}, calledEpochs)
}

func TestPruner_PruneEntityType_NotDoneStops(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	service := &Service{
		cfg: Config{
			Repo:                     mockRepo,
			Metrics:                  mockMetrics,
			SignatureRetentionEpochs: 4,
			PruneBatchSize:           10,
		},
	}

	// batchFunc returns done=false on epoch 5, should stop after one call
	callCount := 0
	service.pruneEntityType(
		context.Background(),
		symbiotic.Epoch(10), symbiotic.Epoch(5),
		4, "test",
		func(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
			callCount++
			require.Equal(t, symbiotic.Epoch(5), epoch)
			return false, nil
		},
	)

	require.Equal(t, 1, callCount)
}

func TestPruner_RetentionZeroSkips(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	service := &Service{
		cfg: Config{
			Repo:    mockRepo,
			Metrics: mockMetrics,
		},
	}

	service.pruneEntityType(
		context.Background(),
		symbiotic.Epoch(100), symbiotic.Epoch(0),
		0, "test",
		func(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
			t.Fatal("should not be called with retention=0")
			return false, nil
		},
	)
}

func TestPruner_ValsetPruneAlwaysDone(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	ctx := context.Background()

	service := &Service{
		cfg: Config{
			Repo:    mockRepo,
			Metrics: mockMetrics,
		},
	}

	mockRepo.EXPECT().PruneValsetEntities(gomock.Any(), symbiotic.Epoch(5), 0).Return(nil)

	done, err := service.pruneValsetEpoch(ctx, symbiotic.Epoch(5))
	require.NoError(t, err)
	require.True(t, done)
}

func TestPruner_IndexBatch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)
	ctx := context.Background()

	batchSize := 5
	requestIDs := []common.Hash{common.HexToHash("0x01"), common.HexToHash("0x02")}

	service := &Service{
		cfg: Config{
			Repo:           mockRepo,
			Metrics:        mockMetrics,
			PruneBatchSize: batchSize,
		},
	}

	// Tick 1: has requestIDs
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(requestIDs, nil)
	mockRepo.EXPECT().PruneEpochIndicesByRequestIDs(gomock.Any(), symbiotic.Epoch(0), requestIDs, batchSize).Return(nil)

	done, err := service.pruneIndexBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.False(t, done)

	// Tick 2: empty
	mockRepo.EXPECT().GetRequestIDsByEpoch(gomock.Any(), symbiotic.Epoch(0), batchSize).Return(nil, nil)

	done, err = service.pruneIndexBatch(ctx, symbiotic.Epoch(0))
	require.NoError(t, err)
	require.True(t, done)
}
