package bbolt

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

// fuzz_buildEpochCorpus seeds a deterministic corpus into the repo:
//   - epoch=100: numReqs signature requests + matching aggregation proofs and
//     sigsPerGroup signatures per request.
//   - epoch=200: a small decoy set to verify epoch isolation.
//
// Returns the main-epoch requestIDs in storage order plus the total number of
// signatures actually written.
func fuzz_buildEpochCorpus(t *testing.T, repo *Repository, numReqs, sigsPerGroup int) (mainReqIDs []common.Hash, totalSigs int) {
	t.Helper()
	ctx := context.Background()

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	mainEpoch := symbiotic.Epoch(100)
	decoyEpoch := symbiotic.Epoch(200)

	for i := 0; i < numReqs; i++ {
		msg := make([]byte, 32)
		_, _ = rand.Read(msg)

		req := symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: mainEpoch, Message: msg}
		sig := symbiotic.Signature{MessageHash: msg, KeyTag: 15, Epoch: mainEpoch, PublicKey: priv.PublicKey()}
		reqID := sig.RequestID()
		mainReqIDs = append(mainReqIDs, reqID)

		require.NoError(t, repo.SaveSignatureRequest(ctx, reqID, req))
		require.NoError(t, repo.saveAggregationProof(ctx, reqID, symbiotic.AggregationProof{MessageHash: msg, KeyTag: 15, Epoch: mainEpoch, Proof: msg}))
		for v := 0; v < sigsPerGroup; v++ {
			require.NoError(t, repo.saveSignatureWithPending(ctx, uint32(v), sig))
			totalSigs++
		}
	}

	// Decoy epoch (verifies isolation).
	for i := 0; i < 5; i++ {
		msg := make([]byte, 32)
		_, _ = rand.Read(msg)
		req := symbiotic.SignatureRequest{KeyTag: 15, RequiredEpoch: decoyEpoch, Message: msg}
		sig := symbiotic.Signature{MessageHash: msg, KeyTag: 15, Epoch: decoyEpoch, PublicKey: priv.PublicKey()}
		require.NoError(t, repo.SaveSignatureRequest(ctx, sig.RequestID(), req))
	}

	// bbolt sorts by raw bytes; for fixed-length 32-byte values lex hex == lex bytes.
	sort.Slice(mainReqIDs, func(i, j int) bool { return mainReqIDs[i].Hex() < mainReqIDs[j].Hex() })
	return mainReqIDs, totalSigs
}

// allPages drains every page of a paginated callable into a flat slice and the
// list of cursors handed back. Asserts no infinite loop.
func allPages[T any](t *testing.T, pageSize int, fn func(from []byte) ([]T, []byte, error)) ([]T, [][]byte) {
	t.Helper()
	var (
		all     []T
		from    []byte
		cursors [][]byte
		safety  = 100000
	)
	for {
		page, next, err := fn(from)
		require.NoError(t, err)
		all = append(all, page...)
		if next == nil {
			break
		}
		require.LessOrEqual(t, len(page), pageSize, "page exceeded pageSize")
		cursors = append(cursors, next)
		from = next
		safety--
		require.Positive(t, safety, "pagination did not terminate")
	}
	return all, cursors
}

func TestFuzz_GetSignatureRequestsWithIDByEpoch_Invariants(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const numReqs = 200
	mainReqIDs, _ := fuzz_buildEpochCorpus(t, repo, numReqs, 0)

	// 1. unbounded baseline.
	allReqs, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Nil(t, next, "unbounded must return nil cursor")
	require.Len(t, allReqs, numReqs)
	for i, r := range allReqs {
		require.Equal(t, mainReqIDs[i], r.RequestID, "ordering at index %d", i)
	}

	// 2. various pageSizes — union must equal baseline, no dups.
	for _, ps := range []int{1, 7, 13, 50, numReqs - 1, numReqs, numReqs + 5, 1000} {
		t.Run("pageSize="+itoa(ps), func(t *testing.T) {
			gathered, _ := allPages(t, ps, func(from []byte) ([]entity.SignatureRequestWithID, []byte, error) {
				return repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, ps, from)
			})
			require.Len(t, gathered, numReqs, "ps=%d total mismatch", ps)
			seen := map[common.Hash]int{}
			for i, r := range gathered {
				seen[r.RequestID]++
				require.Equal(t, mainReqIDs[i], r.RequestID, "ps=%d order broken at %d", ps, i)
			}
			for h, c := range seen {
				require.Equal(t, 1, c, "duplicate %s ps=%d", h.Hex(), ps)
			}
		})
	}

	// 3. exact pageSize == numReqs → 1 page, next == nil.
	page, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, numReqs, nil)
	require.NoError(t, err)
	require.Len(t, page, numReqs)
	require.Nil(t, next, "page filled exactly to numReqs must signal last-page (next=nil)")

	// 4. pageSize == numReqs - 1 → 1st page = numReqs-1, 2nd page = 1, then next==nil.
	first, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, numReqs-1, nil)
	require.NoError(t, err)
	require.Len(t, first, numReqs-1)
	require.NotNil(t, next)
	second, next, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, numReqs-1, next)
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Nil(t, next)
	require.Equal(t, mainReqIDs[numReqs-1], second[0].RequestID)

	// 5. cursor exclusivity — taking lastID as cursor must NOT return that item.
	first2, next2, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, nil)
	require.NoError(t, err)
	require.Len(t, first2, 10)
	require.NotNil(t, next2)
	require.Equal(t, mainReqIDs[9].Bytes(), next2, "cursor must be last-returned ID's raw bytes")

	second2, _, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, next2)
	require.NoError(t, err)
	for _, r := range second2 {
		require.NotEqual(t, mainReqIDs[9], r.RequestID, "exclusive cursor regression")
	}
	require.Equal(t, mainReqIDs[10], second2[0].RequestID)

	// 6. cursor on non-existent hash (random; gap between two real ones).
	randomHash := common.Hash{}
	_, _ = rand.Read(randomHash[:])
	notExisting := randomHash
	for _, h := range mainReqIDs {
		if h == notExisting {
			notExisting = common.Hash{1}
			break
		}
	}
	page3, _, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, numReqs, notExisting.Bytes())
	require.NoError(t, err)
	// Result is real items strictly greater than `notExisting`. No assert on count, just that no item == notExisting.
	for _, r := range page3 {
		require.NotEqual(t, notExisting, r.RequestID)
	}

	// 7. cursor past the very end → empty page, next==nil.
	maxHash := common.Hash{}
	for i := range maxHash {
		maxHash[i] = 0xff
	}
	pageEnd, nextEnd, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, maxHash.Bytes())
	require.NoError(t, err)
	require.Empty(t, pageEnd)
	require.Nil(t, nextEnd)

	// 8. invalid cursors — wrong length.
	for _, badLen := range []int{1, 31, 33, 35, 64} {
		bad := make([]byte, badLen)
		_, _, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 100, 10, bad)
		require.Error(t, err, "expected ErrInvalidCursor for len %d", badLen)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "len %d", badLen)
	}

	// 9. empty epoch → empty page, nil next.
	emptyPage, emptyNext, err := repo.GetSignatureRequestsWithIDByEpoch(ctx, 999, 10, nil)
	require.NoError(t, err)
	require.Empty(t, emptyPage)
	require.Nil(t, emptyNext)
}

func TestFuzz_GetSignatureRequestIDsByEpoch_MatchesRequests(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const numReqs = 150
	mainReqIDs, _ := fuzz_buildEpochCorpus(t, repo, numReqs, 0)

	for _, ps := range []int{1, 50, numReqs} {
		t.Run("pageSize="+itoa(ps), func(t *testing.T) {
			ids, _ := allPages(t, ps, func(from []byte) ([]common.Hash, []byte, error) {
				return repo.GetSignatureRequestIDsByEpoch(ctx, 100, ps, from)
			})
			require.Len(t, ids, numReqs)
			for i, h := range ids {
				require.Equal(t, mainReqIDs[i], h, "ordering at %d ps=%d", i, ps)
			}
		})
	}
}

func TestFuzz_GetAggregationProofsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const numReqs = 100
	mainProofIDs, _ := fuzz_buildEpochCorpus(t, repo, numReqs, 0)

	allProofs, _, err := repo.GetAggregationProofsByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Len(t, allProofs, numReqs)
	for i, p := range allProofs {
		require.Equal(t, mainProofIDs[i], p.RequestID(), "order at %d", i)
	}

	// Page through and reconstruct.
	gathered, _ := allPages(t, 7, func(from []byte) ([]symbiotic.AggregationProof, []byte, error) {
		return repo.GetAggregationProofsByEpoch(ctx, 100, 7, from)
	})
	require.Len(t, gathered, numReqs)
	for i, p := range gathered {
		require.Equal(t, mainProofIDs[i], p.RequestID(), "paginated order at %d", i)
	}

	// Invalid cursor lengths.
	for _, l := range []int{1, 31, 33, 36} {
		bad := make([]byte, l)
		_, _, err := repo.GetAggregationProofsByEpoch(ctx, 100, 5, bad)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "len %d", l)
	}
}

func TestFuzz_GetSignaturesByEpoch_PerSignaturePagination(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := context.Background()

	const (
		numReqs       = 30
		sigsPerGroup  = 7
		expectedTotal = numReqs * sigsPerGroup
	)
	_, total := fuzz_buildEpochCorpus(t, repo, numReqs, sigsPerGroup)
	require.Equal(t, expectedTotal, total)

	// 1. Unbounded baseline.
	all, next, err := repo.GetSignaturesByEpoch(ctx, 100, 0, nil)
	require.NoError(t, err)
	require.Nil(t, next)
	require.Len(t, all, expectedTotal)

	// 2. Reconstruct via various pageSizes — including ones that cut groups in
	//    half (sigsPerGroup-2, sigsPerGroup-1, sigsPerGroup, sigsPerGroup+1).
	for _, ps := range []int{1, 2, sigsPerGroup - 1, sigsPerGroup, sigsPerGroup + 1, 50, expectedTotal - 1, expectedTotal, expectedTotal + 1} {
		t.Run("pageSize="+itoa(ps), func(t *testing.T) {
			gathered, _ := allPages(t, ps, func(from []byte) ([]symbiotic.Signature, []byte, error) {
				return repo.GetSignaturesByEpoch(ctx, 100, ps, from)
			})
			require.Len(t, gathered, expectedTotal, "ps=%d", ps)
			// Compare set: same MessageHash+vIdx? Since we don't know vIdx from
			// Signature struct, we instead verify each baseline element appears
			// the same number of times.
			countAll := signatureBag(all)
			countGathered := signatureBag(gathered)
			require.Equal(t, countAll, countGathered, "ps=%d bag mismatch", ps)
		})
	}

	// 3. exact pageSize == expectedTotal → 1 page, next==nil.
	page, next, err := repo.GetSignaturesByEpoch(ctx, 100, expectedTotal, nil)
	require.NoError(t, err)
	require.Len(t, page, expectedTotal)
	require.Nil(t, next, "exact total page must signal last")

	// 4. invalid cursors for sig endpoint (must be 36 bytes).
	for _, l := range []int{1, 31, 32, 35, 37, 64} {
		bad := make([]byte, l)
		_, _, err := repo.GetSignaturesByEpoch(ctx, 100, 5, bad)
		require.ErrorIs(t, err, entity.ErrInvalidCursor, "sig cursor len %d must be invalid", l)
	}

	// 5. boundary: pageSize=1 with sigsPerGroup=7 → 7 pages traverse one group then move.
	gathered1, _ := allPages(t, 1, func(from []byte) ([]symbiotic.Signature, []byte, error) {
		return repo.GetSignaturesByEpoch(ctx, 100, 1, from)
	})
	require.Len(t, gathered1, expectedTotal)

	// 6. cursor bytes layout: cursor === requestID || BE32(vIdx). Construct
	//    a synthetic cursor pointing right BEFORE the first signature of the
	//    very first group: requestID=mainReqIDs[0], vIdx=MaxUint32 (overflow → 0
	//    is filtered by `+1` in code). The first call without cursor should
	//    start at vIdx=0; an explicit cursor (mainReqIDs[0], MaxUint32) tests
	//    overflow safety.
	mainReqIDs, _ := fuzz_buildEpochCorpus(t, setupTestRepository(t), numReqs, sigsPerGroup)
	curOverflow := make([]byte, 36)
	copy(curOverflow, mainReqIDs[0].Bytes())
	binary.BigEndian.PutUint32(curOverflow[32:], 0xFFFFFFFF)
	_, _, err = repo.GetSignaturesByEpoch(ctx, 100, 5, curOverflow)
	require.NoError(t, err, "MaxUint32 vIdx cursor must not crash (overflow check)")
}

// signatureBag counts each unique (MessageHash, KeyTag, Epoch, PubKey) signature.
func signatureBag(sigs []symbiotic.Signature) map[string]int {
	bag := map[string]int{}
	for _, s := range sigs {
		key := string(s.MessageHash) + "|" + itoa(int(s.KeyTag)) + "|" + itoa(int(s.Epoch)) + "|" + string(s.PublicKey.Raw())
		bag[key]++
	}
	return bag
}

func TestFuzz_RepoutilCursor_RoundTrip(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		var h common.Hash
		_, _ = rand.Read(h[:])
		raw := repoutil.EncodeHashCursor(h)
		require.Len(t, raw, 32)
		got, err := repoutil.DecodeHashCursor(raw)
		require.NoError(t, err)
		require.Equal(t, h, got)
	}
	// signature cursor round trip
	for i := 0; i < 100; i++ {
		var h common.Hash
		_, _ = rand.Read(h[:])
		v := uint32(i * 31)
		raw := repoutil.EncodeSignatureCursor(h, v)
		require.Len(t, raw, 36)
		gotH, gotV, err := repoutil.DecodeSignatureCursor(raw)
		require.NoError(t, err)
		require.Equal(t, h, gotH)
		require.Equal(t, v, gotV)
	}

	// nil/empty
	h, err := repoutil.DecodeHashCursor(nil)
	require.NoError(t, err)
	require.Equal(t, common.Hash{}, h)
	hh, vv, err := repoutil.DecodeSignatureCursor(nil)
	require.NoError(t, err)
	require.Equal(t, common.Hash{}, hh)
	require.Equal(t, uint32(0), vv)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
