package api_server

import (
	"encoding/base64"

	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
)

// Server-side defaults and clamps for *ByEpoch listing endpoints. These bound
// memory usage per response; clients can request a smaller page_size or rely
// on the default. Anything above the max is silently clamped.
const (
	// For ID-only listings (a hex string per item is cheap to materialize).
	defaultIDListPageSize = 1000
	maxIDListPageSize     = 10000

	// For full-object listings (signatures, proofs, signature requests).
	defaultListPageSize = 100
	maxListPageSize     = 1000
)

// clampPageSize returns the effective page size: 0 → defaultV, > maxV → maxV.
func clampPageSize(requested int, defaultV, maxV int) int {
	switch {
	case requested <= 0:
		return defaultV
	case requested > maxV:
		return maxV
	default:
		return requested
	}
}

// decodeCursor parses an opaque base64 cursor returned by a previous response.
// Empty string yields nil (signals "start of range"). Malformed base64 yields
// codes.InvalidArgument so the client gets a useful error.
func decodeCursor(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid `from`: %v", err)
	}
	return raw, nil
}

// encodeCursor serializes a repo-supplied opaque cursor back into the wire
// format. Nil/empty input means "this is the last page".
func encodeCursor(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// asCursorErr returns a non-nil InvalidArgument status when err wraps
// entity.ErrInvalidCursor, and nil otherwise — letting callers fall through to
// their own wrapping for unrelated repo errors.
func asCursorErr(err error) error {
	if errors.Is(err, entity.ErrInvalidCursor) {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return nil
}
