package pruner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/symbioticfi/relay/internal/usecase/pruner/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestPruner_RetentionCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		retention           uint64
		latestEpoch         symbiotic.Epoch
		oldestStoredEpoch   symbiotic.Epoch
		expectedPruneEpochs []symbiotic.Epoch
		expectedNoPrune     bool
	}{
		{
			name:                "retention=4, latestEpoch=10 should keep exactly 4 epochs (7-10)",
			retention:           4,
			latestEpoch:         10,
			oldestStoredEpoch:   0,
			expectedPruneEpochs: []symbiotic.Epoch{0, 1, 2, 3, 4, 5, 6},
			expectedNoPrune:     false,
		},
		{
			name:                "retention=4, latestEpoch=4 should keep exactly 4 epochs (1-4)",
			retention:           4,
			latestEpoch:         4,
			oldestStoredEpoch:   0,
			expectedPruneEpochs: []symbiotic.Epoch{0},
			expectedNoPrune:     false,
		},
		{
			name:                "retention=4, latestEpoch=3 should not prune (less than retention)",
			retention:           4,
			latestEpoch:         3,
			oldestStoredEpoch:   0,
			expectedPruneEpochs: nil,
			expectedNoPrune:     true,
		},
		{
			name:                "retention=1, latestEpoch=5 should keep only latest epoch",
			retention:           1,
			latestEpoch:         5,
			oldestStoredEpoch:   0,
			expectedPruneEpochs: []symbiotic.Epoch{0, 1, 2, 3, 4},
			expectedNoPrune:     false,
		},
		{
			name:                "retention=10, latestEpoch=100 should keep exactly 10 epochs (91-100)",
			retention:           10,
			latestEpoch:         100,
			oldestStoredEpoch:   0,
			expectedPruneEpochs: makeRange(0, 90),
			expectedNoPrune:     false,
		},
		{
			name:                "oldestStoredEpoch already within retention window",
			retention:           4,
			latestEpoch:         10,
			oldestStoredEpoch:   7,
			expectedPruneEpochs: nil,
			expectedNoPrune:     true,
		},
		{
			name:                "oldestStoredEpoch partially in pruning range",
			retention:           4,
			latestEpoch:         10,
			oldestStoredEpoch:   5,
			expectedPruneEpochs: []symbiotic.Epoch{5, 6},
			expectedNoPrune:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockrepo(ctrl)
			mockMetrics := mocks.NewMockmetrics(ctrl)

			ctx := context.Background()

			// Set expectations: expect calls for each epoch to be pruned
			if !tt.expectedNoPrune {
				for _, epoch := range tt.expectedPruneEpochs {
					mockRepo.EXPECT().PruneValsetEntities(gomock.Any(), epoch).Return(nil)
					mockMetrics.EXPECT().IncPrunedEpochsCount("valset")
				}
			}

			service := &Service{
				cfg: Config{
					Repo:                  mockRepo,
					Metrics:               mockMetrics,
					ValsetRetentionEpochs: tt.retention,
				},
			}

			count, err := service.pruneValsetEntities(ctx, tt.latestEpoch, tt.oldestStoredEpoch)
			require.NoError(t, err)

			expectedCount := uint64(len(tt.expectedPruneEpochs))
			require.Equal(t, expectedCount, count, "pruned epoch count mismatch")
		})
	}
}

func TestPruner_RetentionCalculation_AllEntityTypes(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	ctx := context.Background()
	retention := uint64(4)
	latestEpoch := symbiotic.Epoch(10)
	oldestStoredEpoch := symbiotic.Epoch(0)

	// Expected to prune epochs 0-6, keep 7-10 (4 epochs)
	expectedPruneEpochs := []symbiotic.Epoch{0, 1, 2, 3, 4, 5, 6}

	t.Run("valset entities", func(t *testing.T) {
		for _, epoch := range expectedPruneEpochs {
			mockRepo.EXPECT().PruneValsetEntities(gomock.Any(), epoch).Return(nil)
			mockMetrics.EXPECT().IncPrunedEpochsCount("valset")
		}

		service := &Service{
			cfg: Config{
				Repo:                  mockRepo,
				Metrics:               mockMetrics,
				ValsetRetentionEpochs: retention,
			},
		}

		count, err := service.pruneValsetEntities(ctx, latestEpoch, oldestStoredEpoch)
		require.NoError(t, err)
		require.Equal(t, uint64(7), count)
	})

	t.Run("proof entities", func(t *testing.T) {
		for _, epoch := range expectedPruneEpochs {
			mockRepo.EXPECT().PruneProofEntities(gomock.Any(), epoch).Return(nil)
			mockMetrics.EXPECT().IncPrunedEpochsCount("proof")
		}

		service := &Service{
			cfg: Config{
				Repo:                 mockRepo,
				Metrics:              mockMetrics,
				ProofRetentionEpochs: retention,
			},
		}

		count, err := service.pruneProofEntities(ctx, latestEpoch, oldestStoredEpoch)
		require.NoError(t, err)
		require.Equal(t, uint64(7), count)
	})

	t.Run("signature entities", func(t *testing.T) {
		for _, epoch := range expectedPruneEpochs {
			mockRepo.EXPECT().PruneSignatureEntitiesForEpoch(gomock.Any(), epoch).Return(nil)
			mockMetrics.EXPECT().IncPrunedEpochsCount("signature")
		}

		service := &Service{
			cfg: Config{
				Repo:                     mockRepo,
				Metrics:                  mockMetrics,
				SignatureRetentionEpochs: retention,
			},
		}

		count, err := service.pruneSignatureEntities(ctx, latestEpoch, oldestStoredEpoch)
		require.NoError(t, err)
		require.Equal(t, uint64(7), count)
	})
}

func TestPruner_EdgeCases(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockrepo(ctrl)
	mockMetrics := mocks.NewMockmetrics(ctrl)

	ctx := context.Background()

	t.Run("retention=0 should not prune", func(t *testing.T) {
		service := &Service{
			cfg: Config{
				Repo:                  mockRepo,
				Metrics:               mockMetrics,
				ValsetRetentionEpochs: 0,
			},
		}

		count, err := service.pruneValsetEntities(ctx, 100, 0)
		require.NoError(t, err)
		require.Equal(t, uint64(0), count)
	})

	t.Run("latestEpoch equals retention should keep all epochs starting from 1", func(t *testing.T) {
		retention := uint64(5)
		latestEpoch := symbiotic.Epoch(5)
		oldestStoredEpoch := symbiotic.Epoch(0)

		// Should prune epoch 0, keep 1-5 (5 epochs)
		mockRepo.EXPECT().PruneValsetEntities(gomock.Any(), symbiotic.Epoch(0)).Return(nil)
		mockMetrics.EXPECT().IncPrunedEpochsCount("valset")

		service := &Service{
			cfg: Config{
				Repo:                  mockRepo,
				Metrics:               mockMetrics,
				ValsetRetentionEpochs: retention,
			},
		}

		count, err := service.pruneValsetEntities(ctx, latestEpoch, oldestStoredEpoch)
		require.NoError(t, err)
		require.Equal(t, uint64(1), count)
	})
}

// Helper function to generate a range of epochs
func makeRange(start, end symbiotic.Epoch) []symbiotic.Epoch {
	result := make([]symbiotic.Epoch, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
}
