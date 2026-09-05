package sync_provider

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type regressionMissingRepo struct {
	repo

	lookups int
}

func (r *regressionMissingRepo) GetSignatureByIndex(context.Context, common.Hash, uint32) (symbiotic.Signature, error) {
	r.lookups++
	return symbiotic.Signature{}, entity.ErrEntityNotFound
}
func TestRegressionMissingBitmapWork(t *testing.T) {
	bitmap := entity.NewBitmap()
	bitmap.AddRange(0, 1000000)
	bitmap.RunOptimize()
	encoded, err := bitmap.ToBytes()
	require.NoError(t, err)
	repository := &regressionMissingRepo{}
	s := &Syncer{cfg: Config{Repo: repository, MaxResponseSignatureCount: 1, MaxSignatureRequestsPerSync: 1}}
	_, err = s.HandleWantSignaturesRequest(t.Context(), entity.WantSignaturesRequest{WantSignatures: map[common.Hash]entity.Bitmap{{}: bitmap}})
	require.NoError(t, err)
	t.Logf("bitmapBytes=%d cardinality=%d responseLimit=1 databaseLookups=%d", len(encoded), bitmap.GetCardinality(), repository.lookups)
	require.Equal(t, 1, repository.lookups)
}

func TestWantSignaturesRejectsTooManyRequests(t *testing.T) {
	repository := &regressionMissingRepo{}
	s := &Syncer{cfg: Config{Repo: repository, MaxResponseSignatureCount: 10, MaxSignatureRequestsPerSync: 1}}
	_, err := s.HandleWantSignaturesRequest(t.Context(), entity.WantSignaturesRequest{
		WantSignatures: map[common.Hash]entity.Bitmap{
			{1}: entity.NewBitmap(),
			{2}: entity.NewBitmap(),
		},
	})
	require.ErrorContains(t, err, "too many")
	require.Zero(t, repository.lookups)
}

func TestWantSignaturesHonorsCancellationBeforeLookup(t *testing.T) {
	repository := &regressionMissingRepo{}
	s := &Syncer{cfg: Config{Repo: repository, MaxResponseSignatureCount: 10, MaxSignatureRequestsPerSync: 1}}
	bitmap := entity.NewBitmap()
	bitmap.Add(1)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := s.HandleWantSignaturesRequest(ctx, entity.WantSignaturesRequest{
		WantSignatures: map[common.Hash]entity.Bitmap{{1}: bitmap},
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, repository.lookups)
}
