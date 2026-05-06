package bbolt

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
)

// Internal pagination cursor format. Public API exposes opaque base64; the
// repo serializes raw bytes here. Two layouts are used:
//
//   - hash cursor (32 bytes): requestID. Used by Requests/IDs/Proofs listings.
//   - signature cursor (36 bytes): requestID || BE32(validatorIndex). Used by
//     Signatures listing because pagination is per-signature within a group.
const signatureCursorLen = common.HashLength + 4

func decodeHashCursor(from []byte) (common.Hash, error) {
	if len(from) == 0 {
		return common.Hash{}, nil
	}
	if len(from) != common.HashLength {
		return common.Hash{}, errors.Errorf("%w: expected %d bytes, got %d", entity.ErrInvalidCursor, common.HashLength, len(from))
	}
	return common.BytesToHash(from), nil
}

func encodeHashCursor(h common.Hash) []byte {
	return h.Bytes()
}

func decodeSignatureCursor(from []byte) (common.Hash, uint32, error) {
	if len(from) == 0 {
		return common.Hash{}, 0, nil
	}
	if len(from) != signatureCursorLen {
		return common.Hash{}, 0, errors.Errorf("%w: expected %d bytes, got %d", entity.ErrInvalidCursor, signatureCursorLen, len(from))
	}
	return common.BytesToHash(from[:common.HashLength]), binary.BigEndian.Uint32(from[common.HashLength:]), nil
}

func encodeSignatureCursor(h common.Hash, vIdx uint32) []byte {
	buf := make([]byte, signatureCursorLen)
	copy(buf, h.Bytes())
	binary.BigEndian.PutUint32(buf[common.HashLength:], vIdx)
	return buf
}
