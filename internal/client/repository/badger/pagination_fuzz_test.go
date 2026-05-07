package badger

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"sort"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

// fuzz_buildEpochCorpus seeds a deterministic corpus into the repo:
// epoch=100 with numReqs requests, sigsPerGroup signatures per request,
// matching aggregation proofs; plus a small decoy in epoch=200.
func fuzz_buildEpochCorpus(t *testing.T, repo *Repository, numReqs, sigsPerGroup int) (mainIDs []common.Hash, totalSigs int) {
	t.Helper()
	ctx := context.Background()

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	mainEpoch := symbiotic.Epoch(100)
	for i := 0; i < numReqs; i++ {
		msg := make([]byte, 32)
		_, _ = rand.Read(msg)
		req := symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: mainEpoch, Message: msg}
		sig := symbiotic.Signature{MessageHash: msg, KeyTag: 15, Epoch: mainEpoch, PublicKey: priv.PublicKey()}
		reqID := sig.RequestID()
		mainIDs = append(mainIDs, reqID)

		require.NoError(t, repo.SaveSignatureRequest(ctx, reqID, req))
		require.NoError(t, repo.saveAggregationProof(ctx, reqID, symbiotic.AggregationProof{MessageHash: msg, KeyTag: 15, Epoch: mainEpoch, Proof: msg}))
		for v := 0; v < sigsPerGroup; v++ {
			require.NoError(t, repo.saveSignature(ctx, uint32(v), sig))
			totalSigs++
		}
	}

	// Decoy epoch.
	decoyEpoch := symbiotic.Epoch(200)
	for i := 0; i < 5; i++ {
		msg := make([]byte, 32)
		_, _ = rand.Read(msg)
		req := symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: decoyEpoch, Message: msg}
		sig := symbiotic.Signature{MessageHash: msg, KeyTag: 15, Epoch: decoyEpoch, PublicKey: priv.PublicKey()}
		require.NoError(t, repo.SaveSignatureRequest(ctx, sig.RequestID(), req))
	}

	// badger keys store hash.Hex() — lex hex order = lex byte order for fixed-length values.
	sort.Slice(mainIDs, func(i, j int) bool { return mainIDs[i].Hex() < mainIDs[j].Hex() })
	return mainIDs, totalSigs
}

func allPages[T any](t *testing.T, pageSize int, fn func(from []byte) ([]T, []byte, error)) []T {
	t.Helper()
	var (
		all    []T
		from   []byte
		safety = 100000
	)
	for {
		page, next, err := fn(from)
		require.NoError(t, err)
		all = append(all, page...)
		if next == nil {
			break
		}
		require.LessOrEqual(t, len(page), pageSize, "page exceeded pageSize")
		from = next
		safety--
		require.Positive(t, safety)
	}
	return all
}

func TestFuzz_Badger_RequestsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const numReqs = 200
	mainIDs, _ := fuzz_buildEpochCorpus(t, repo, numReqs, 0)

	// 1. unbounded.
	all, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Nil(t, next)
	require.Len(t, all, numReqs)
	for i, r := range all {
		require.Equal(t, mainIDs[i], r.RequestID, "order at %d", i)
	}

	// 2. various pageSizes.
	for _, ps := range []int{1, 7, 50, numReqs - 1, numReqs, numReqs + 5} {
		t.Run("pageSize="+strconv.Itoa(ps), func(t *testing.T) {
			gathered := allPages(t, ps, func(from []byte) ([]entity.SignatureRequestWithID, []byte, error) {
				return repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, ps, from)
			})
			require.Len(t, gathered, numReqs)
			for i, r := range gathered {
				require.Equal(t, mainIDs[i], r.RequestID, "order at %d ps=%d", i, ps)
			}
		})
	}

	// 3. exact total → next == nil.
	page, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, numReqs, nil)
	require.NoError(t, err)
	require.Len(t, page, numReqs)
	require.Nil(t, next, "exact total page must signal last")

	// 4. cursor exclusivity.
	first, next2, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, nil)
	require.NoError(t, err)
	require.Len(t, first, 10)
	require.NotNil(t, next2)
	require.Equal(t, mainIDs[9].Bytes(), next2)

	second, _, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, next2)
	require.NoError(t, err)
	for _, r := range second {
		require.NotEqual(t, mainIDs[9], r.RequestID)
	}
	require.Equal(t, mainIDs[10], second[0].RequestID)

	// 5. invalid cursor lengths.
	for _, l := range []int{1, 31, 33, 35} {
		bad := make([]byte, l)
		_, _, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 5, bad)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "len %d", l)
	}

	// 6. cursor past the end.
	maxHash := common.Hash{}
	for i := range maxHash {
		maxHash[i] = 0xff
	}
	pageEnd, nextEnd, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, maxHash.Bytes())
	require.NoError(t, err)
	require.Empty(t, pageEnd)
	require.Nil(t, nextEnd)

	// 7. empty epoch.
	empty, emptyNext, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 999, 10, nil)
	require.NoError(t, err)
	require.Empty(t, empty)
	require.Nil(t, emptyNext)
}

func TestFuzz_Badger_SignaturesByEpoch_PerSignature(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const (
		numReqs       = 30
		sigsPerGroup  = 7
		expectedTotal = numReqs * sigsPerGroup
	)
	mainIDs, total := fuzz_buildEpochCorpus(t, repo, numReqs, sigsPerGroup)
	require.Equal(t, expectedTotal, total)

	// Unbounded baseline.
	all, next, err := repo.GetSignaturesByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Nil(t, next)
	require.Len(t, all, expectedTotal)

	// Reconstruct via various pageSizes including ones that cut groups.
	for _, ps := range []int{1, 2, sigsPerGroup - 1, sigsPerGroup, sigsPerGroup + 1, 50, expectedTotal - 1, expectedTotal} {
		t.Run("pageSize="+strconv.Itoa(ps), func(t *testing.T) {
			gathered := allPages(t, ps, func(from []byte) ([]symbiotic.Signature, []byte, error) {
				return repo.GetSignaturesByEpoch(ctx, 100, ps, from)
			})
			require.Len(t, gathered, expectedTotal, "ps=%d", ps)
		})
	}

	// MaxUint32 cursor must not crash.
	curOverflow := make([]byte, 36)
	copy(curOverflow, mainIDs[0].Bytes())
	binary.BigEndian.PutUint32(curOverflow[32:], 0xFFFFFFFF)
	_, _, err = repo.GetSignaturesByEpoch(ctx, 100, 5, curOverflow)
	require.NoError(t, err, "MaxUint32 vIdx cursor must not crash")

	// Invalid lengths.
	for _, l := range []int{1, 31, 32, 35, 37} {
		bad := make([]byte, l)
		_, _, err := repo.GetSignaturesByEpoch(ctx, 100, 5, bad)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "len %d", l)
	}

	// Cursor at exactly last vIdx of last group → empty next page.
	cursorLast := repoutil.EncodeSignatureCursor(mainIDs[len(mainIDs)-1], uint32(sigsPerGroup-1))
	pageAfterEnd, nextAfterEnd, err := repo.GetSignaturesByEpoch(ctx, 100, 10, cursorLast)
	require.NoError(t, err)
	require.Empty(t, pageAfterEnd)
	require.Nil(t, nextAfterEnd)
}

func TestFuzz_Badger_AggregationProofsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const numReqs = 100
	mainIDs, _ := fuzz_buildEpochCorpus(t, repo, numReqs, 0)

	all, _, err := repo.GetAggregationProofsByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Len(t, all, numReqs)
	for i, p := range all {
		require.Equal(t, mainIDs[i], p.RequestID(), "order at %d", i)
	}

	// Reconstruct with several pageSizes.
	for _, ps := range []int{1, 7, 50, numReqs} {
		gathered := allPages(t, ps, func(from []byte) ([]symbiotic.AggregationProof, []byte, error) {
			return repo.GetAggregationProofsByEpoch(ctx, 100, ps, from)
		})
		require.Len(t, gathered, numReqs, "ps=%d", ps)
	}

	for _, l := range []int{1, 31, 33, 36} {
		bad := make([]byte, l)
		_, _, err := repo.GetAggregationProofsByEpoch(ctx, 100, 5, bad)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "len %d", l)
	}
}
