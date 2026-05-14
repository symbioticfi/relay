package api_server

import (
	"context"
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// TestFuzz_Handler_PageSizeClamps drives every handler with extreme page_size
// values and verifies the clamp/default constants are applied at the boundary
// to the repo. Catches regressions where the handler accidentally trusts the
// user input as-is.
func TestFuzz_Handler_PageSizeClamps(t *testing.T) {
	cases := []struct {
		name        string
		requested   uint32
		idsExpected int // for IDs handler
		objExpected int // for full-object handlers
	}{
		{"zero=default", 0, defaultIDListPageSize, defaultListPageSize},
		{"one=one", 1, 1, 1},
		{"max=max", maxIDListPageSize, maxIDListPageSize, maxListPageSize},
		{"over_ids=clamp", maxIDListPageSize + 1, maxIDListPageSize, maxListPageSize},
		{"way_over=clamp", 1 << 31, maxIDListPageSize, maxListPageSize},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setup := newTestSetup(t)
			ctx := context.Background()
			epoch := symbiotic.Epoch(7)

			// IDs handler — max is maxIDListPageSize.
			setup.mockRepo.EXPECT().GetSignatureRequestIDsByEpoch(ctx, epoch, c.idsExpected, []byte(nil)).Return(nil, []byte(nil), nil)
			_, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, &apiv1.GetSignatureRequestIDsByEpochRequest{Epoch: 7, PageSize: c.requested})
			require.NoError(t, err)

			// Requests handler — max is maxListPageSize.
			setup.mockRepo.EXPECT().GetSignatureRequestsWithIDByEpoch(ctx, epoch, c.objExpected, []byte(nil)).Return(nil, []byte(nil), nil)
			_, err = setup.handler.GetSignatureRequestsByEpoch(ctx, &apiv1.GetSignatureRequestsByEpochRequest{Epoch: 7, PageSize: c.requested})
			require.NoError(t, err)

			// Signatures handler.
			setup.mockRepo.EXPECT().GetSignaturesByEpoch(ctx, epoch, c.objExpected, []byte(nil)).Return(nil, []byte(nil), nil)
			_, err = setup.handler.GetSignaturesByEpoch(ctx, &apiv1.GetSignaturesByEpochRequest{Epoch: 7, PageSize: c.requested})
			require.NoError(t, err)

			// Aggregation proofs.
			setup.mockRepo.EXPECT().GetAggregationProofsByEpoch(ctx, epoch, c.objExpected, []byte(nil)).Return(nil, []byte(nil), nil)
			_, err = setup.handler.GetAggregationProofsByEpoch(ctx, &apiv1.GetAggregationProofsByEpochRequest{Epoch: 7, PageSize: c.requested})
			require.NoError(t, err)
		})
	}
}

// TestFuzz_Handler_CursorRoundTrip verifies the handler base64-encodes the raw
// cursor returned by the repo, and decodes a client-supplied base64 cursor
// back to the same raw bytes when forwarding to the repo.
func TestFuzz_Handler_CursorRoundTrip(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	epoch := symbiotic.Epoch(7)

	// Repo returns a raw 32-byte cursor; handler must base64-encode it in
	// next_from. Then a follow-up request with that base64 in `from` must
	// reach the repo as the same raw bytes.
	rawCursor := common.HexToHash("0xdeadbeef000000000000000000000000000000000000000000000000cafebabe").Bytes()

	setup.mockRepo.EXPECT().
		GetSignatureRequestIDsByEpoch(ctx, epoch, defaultIDListPageSize, []byte(nil)).
		Return([]common.Hash{{0x01}}, rawCursor, nil)
	resp, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, &apiv1.GetSignatureRequestIDsByEpochRequest{Epoch: 7})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetNextFrom(), "handler must return base64 cursor")

	encoded := resp.GetNextFrom()
	decoded, decErr := base64.RawURLEncoding.DecodeString(encoded)
	require.NoError(t, decErr)
	require.Equal(t, rawCursor, decoded, "handler must base64 the raw repo cursor")

	// Round-trip: send `encoded` back as `from`, repo must see `rawCursor`.
	setup.mockRepo.EXPECT().
		GetSignatureRequestIDsByEpoch(ctx, epoch, defaultIDListPageSize, rawCursor).
		Return(nil, []byte(nil), nil)
	_, err = setup.handler.GetSignatureRequestIDsByEpoch(ctx, &apiv1.GetSignatureRequestIDsByEpochRequest{Epoch: 7, From: encoded})
	require.NoError(t, err)
}

// TestFuzz_Handler_InvalidBase64 verifies the handler short-circuits on
// malformed base64 BEFORE calling the repo (mock would fail otherwise).
func TestFuzz_Handler_InvalidBase64(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	// gomock will FAIL the test if the repo is invoked unexpectedly. So the
	// absence of an EXPECT() call here is the assertion: handler must not
	// call repo on bad cursor.
	bad := "not!valid!base64"
	_, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, &apiv1.GetSignatureRequestIDsByEpochRequest{Epoch: 5, From: bad})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	require.Equal(t, codes.InvalidArgument, st.Code(), "bad base64 must yield InvalidArgument")
}

// TestFuzz_Handler_RepoCursorErrPropagates verifies that ErrInvalidCursor from
// the repo is mapped to InvalidArgument by the handler (asCursorErr).
func TestFuzz_Handler_RepoCursorErrPropagates(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	epoch := symbiotic.Epoch(7)

	// Send a syntactically valid base64 cursor that the repo will reject by
	// length. We pass 5 bytes; repo returns ErrInvalidCursor.
	bad := base64.RawURLEncoding.EncodeToString([]byte{1, 2, 3, 4, 5})
	setup.mockRepo.EXPECT().
		GetSignatureRequestIDsByEpoch(ctx, epoch, defaultIDListPageSize, []byte{1, 2, 3, 4, 5}).
		Return(nil, []byte(nil), entity.ErrInvalidCursor)

	_, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, &apiv1.GetSignatureRequestIDsByEpochRequest{Epoch: 7, From: bad})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code(), "ErrInvalidCursor must surface as InvalidArgument")
}

// TestFuzz_Handler_NextFromEmptyOnLastPage: when the repo signals last page
// (nil cursor), the handler must return next_from == "".
func TestFuzz_Handler_NextFromEmptyOnLastPage(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	epoch := symbiotic.Epoch(7)

	setup.mockRepo.EXPECT().
		GetSignaturesByEpoch(ctx, epoch, defaultListPageSize, []byte(nil)).
		Return([]symbiotic.Signature{}, []byte(nil), nil)
	resp, err := setup.handler.GetSignaturesByEpoch(ctx, &apiv1.GetSignaturesByEpochRequest{Epoch: 7})
	require.NoError(t, err)
	require.Empty(t, resp.GetNextFrom())
}

// TestFuzz_Handler_PartialCursorIsOpaque: an attacker passes garbage of the
// correct length 32 — repo accepts it (no ErrInvalidCursor); handler must
// forward it transparently. Used to fuzz that base64 of any length works.
func TestFuzz_Handler_PartialCursorOpaqueLengths(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()
	epoch := symbiotic.Epoch(7)

	for _, bytesLen := range []int{1, 16, 32, 36, 64} {
		t.Run("bytes="+strconv.Itoa(bytesLen), func(t *testing.T) {
			raw := make([]byte, bytesLen)
			for i := range raw {
				raw[i] = byte(i)
			}
			b64 := base64.RawURLEncoding.EncodeToString(raw)

			// Repo will receive `raw` directly; doesn't matter what it
			// returns — we're just verifying the wire layer.
			setup.mockRepo.EXPECT().
				GetSignaturesByEpoch(ctx, epoch, defaultListPageSize, raw).
				Return(nil, []byte(nil), nil)
			_, err := setup.handler.GetSignaturesByEpoch(ctx, &apiv1.GetSignaturesByEpochRequest{Epoch: 7, From: b64})
			require.NoError(t, err)
		})
	}
}
