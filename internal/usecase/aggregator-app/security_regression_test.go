package aggregator_app

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"k8s.io/client-go/util/workqueue"
)

type regressionPendingRepo struct {
	repository

	rows  []symbiotic.SignatureRequestWithID
	scans int
}

func (r *regressionPendingRepo) GetLatestValidatorSetEpoch(context.Context) (symbiotic.Epoch, error) {
	return 1, nil
}
func (r *regressionPendingRepo) GetAggregationProof(context.Context, common.Hash) (symbiotic.AggregationProof, error) {
	return symbiotic.AggregationProof{}, entity.ErrEntityNotFound
}
func (r *regressionPendingRepo) GetSignatureRequestsWithoutAggregationProof(_ context.Context, _ symbiotic.Epoch, limit int, last common.Hash) ([]symbiotic.SignatureRequestWithID, error) {
	r.scans++
	start := int(last.Big().Int64())
	if start >= len(r.rows) {
		return nil, nil
	}
	return r.rows[start:min(start+limit, len(r.rows))], nil
}
func TestRegressionCatchupStarvation(t *testing.T) {
	r := &regressionPendingRepo{}
	for i := 1; i <= 101; i++ {
		r.rows = append(r.rows, symbiotic.SignatureRequestWithID{RequestID: common.BigToHash(big.NewInt(int64(i))), SignatureRequest: symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: 1}})
	}
	a := &AggregatorApp{cfg: Config{Repo: r, MaxPendingRequests: 1000, ProofCatchup: ProofCatchupConfig{Enabled: true, Interval: time.Minute, EpochsToCheck: 1, MaxRequestsPerCycle: 100}}, queue: workqueue.NewTyped[common.Hash]()}
	seenLast := false
	for cycle := 0; cycle < 3; cycle++ {
		require.NoError(t, a.tryAggregateRequestsWithoutProof(t.Context()))
		for a.queue.Len() > 0 {
			id, _ := a.queue.Get()
			seenLast = seenLast || id == r.rows[100].RequestID
			a.queue.Done(id)
		}
		t.Logf("cycle=%d reachedRequest101=%v", cycle+1, seenLast)
	}
	require.True(t, seenLast)
}

func (r *regressionPendingRepo) GetSignatureMap(context.Context, common.Hash) (entity.SignatureMap, error) {
	return entity.SignatureMap{}, entity.ErrEntityNotFound
}
func TestRegressionCatchupNonAggregationBudget(t *testing.T) {
	r := &regressionPendingRepo{}
	for i := 1; i <= 10000; i++ {
		r.rows = append(r.rows, symbiotic.SignatureRequestWithID{RequestID: common.BigToHash(big.NewInt(int64(i))), SignatureRequest: symbiotic.SignatureRequest{KeyTag: 16, RequiredEpoch: 1}})
	}
	a := &AggregatorApp{cfg: Config{Repo: r, MaxPendingRequests: 1000, ProofCatchup: ProofCatchupConfig{Enabled: true, Interval: time.Minute, EpochsToCheck: 1, MaxRequestsPerCycle: 100}}, queue: workqueue.NewTyped[common.Hash]()}
	require.NoError(t, a.tryAggregateRequestsWithoutProof(t.Context()))
	t.Logf("requestBudget=100 databasePages=%d", r.scans)
	require.Equal(t, 10, r.scans)
}
