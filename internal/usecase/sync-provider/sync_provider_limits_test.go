package sync_provider

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type signatureAvailabilityRepo struct {
	repo

	signatureMap entity.SignatureMap
	calls        []uint32
	err          error
}

func (r *signatureAvailabilityRepo) GetSignatureMap(context.Context, common.Hash) (entity.SignatureMap, error) {
	return r.signatureMap, r.err
}
func (r *signatureAvailabilityRepo) GetSignatureByIndex(_ context.Context, _ common.Hash, index uint32) (symbiotic.Signature, error) {
	r.calls = append(r.calls, index)
	if index == 1 || !r.signatureMap.SignedValidatorsBitmap.Contains(index) {
		return symbiotic.Signature{}, entity.ErrEntityNotFound
	}
	return symbiotic.Signature{Epoch: 1}, nil
}

func TestResponseLimitCountsAvailableSignatures(t *testing.T) {
	id := common.Hash{1}
	r := &signatureAvailabilityRepo{signatureMap: entity.NewSignatureMap(id, 1, 6)}
	r.signatureMap.SignedValidatorsBitmap.AddMany([]uint32{1, 2, 3, 5})
	requested := entity.NewBitmap()
	requested.AddRange(0, 8)
	requested.Remove(2)
	s := &Syncer{cfg: Config{Repo: r, MaxResponseSignatureCount: 1}}
	request := entity.WantSignaturesRequest{WantSignatures: map[common.Hash]entity.Bitmap{id: requested}}
	// A nil bitmap must not discard the valid request in the same batch.
	request.WantSignatures[common.Hash{2}] = entity.Bitmap{}

	// Index 1 disappeared after the bitmap read. Index 3 still fills the response;
	// unrequested index 2 and signatures beyond the response limit are not read.
	response, err := s.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)
	require.Len(t, response.Signatures[id], 1)
	require.Equal(t, uint32(3), response.Signatures[id][0].ValidatorIndex)
	require.Equal(t, []uint32{1, 3}, r.calls)
	require.Equal(t, []uint32{1, 2, 3, 5}, r.signatureMap.SignedValidatorsBitmap.ToArray())
	require.Equal(t, uint64(7), requested.GetCardinality())

	r.err = entity.ErrEntityNotFound
	response, err = s.HandleWantSignaturesRequest(t.Context(), request)
	require.NoError(t, err)
	require.Empty(t, response.Signatures)
	r.err = errors.New("database unavailable")
	_, err = s.HandleWantSignaturesRequest(t.Context(), request)
	require.ErrorIs(t, err, r.err)
	require.Equal(t, []uint32{1, 3}, r.calls)
}
