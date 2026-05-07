package repoutil

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

// DecodeHashCursor parses a 32-byte cursor. Empty cursor → zero-hash (start).
// Wrong length → entity.ErrInvalidCursor.
func DecodeHashCursor(from []byte) (common.Hash, error) {
	if len(from) == 0 {
		return common.Hash{}, nil
	}
	if len(from) != common.HashLength {
		return common.Hash{}, errors.Errorf("%w: expected %d bytes, got %d", entity.ErrInvalidCursor, common.HashLength, len(from))
	}
	return common.BytesToHash(from), nil
}

// EncodeHashCursor returns the raw bytes of a hash cursor.
func EncodeHashCursor(h common.Hash) []byte {
	return h.Bytes()
}

// DecodeSignatureCursor parses a 36-byte composite cursor (requestID + vIdx).
// Empty cursor → (zero-hash, 0) meaning "start from the beginning".
func DecodeSignatureCursor(from []byte) (common.Hash, uint32, error) {
	if len(from) == 0 {
		return common.Hash{}, 0, nil
	}
	if len(from) != signatureCursorLen {
		return common.Hash{}, 0, errors.Errorf("%w: expected %d bytes, got %d", entity.ErrInvalidCursor, signatureCursorLen, len(from))
	}
	return common.BytesToHash(from[:common.HashLength]), binary.BigEndian.Uint32(from[common.HashLength:]), nil
}

// EncodeSignatureCursor packs (hash, vIdx) into the 36-byte cursor format.
func EncodeSignatureCursor(h common.Hash, vIdx uint32) []byte {
	buf := make([]byte, signatureCursorLen)
	copy(buf, h.Bytes())
	binary.BigEndian.PutUint32(buf[common.HashLength:], vIdx)
	return buf
}
