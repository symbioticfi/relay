package bbolt

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) saveProofCommitPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdate(ctx, "saveProofCommitPending", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		b := tx.Bucket(bucketAggProofCommits)
		if b.Get(ek) != nil {
			return errors.Errorf("proof commit pending already exists: %w", entity.ErrEntityAlreadyExist)
		}
		return b.Put(ek, requestID.Bytes())
	})
}

func (r *Repository) removeProofCommitPending(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "removeProofCommitPending", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		b := tx.Bucket(bucketAggProofCommits)
		if b.Get(ek) == nil {
			return errors.Errorf("proof commit pending not found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
		}
		return b.Delete(ek)
	})
}

func (r *Repository) GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]symbiotic.ProofCommitKey, error) {
	var keys []symbiotic.ProofCommitKey

	err := r.doView(ctx, "GetPendingProofCommitsSinceEpoch", func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketAggProofCommits).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(k) != 8 || len(v) != 32 {
				continue
			}
			keyEpoch := symbiotic.Epoch(binary.BigEndian.Uint64(k))
			if keyEpoch < epoch {
				continue
			}

			keys = append(keys, symbiotic.ProofCommitKey{
				Epoch:     keyEpoch,
				RequestID: common.BytesToHash(v),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Epoch != keys[j].Epoch {
			return keys[i].Epoch < keys[j].Epoch
		}
		return bytes.Compare(keys[i].RequestID[:], keys[j].RequestID[:]) < 0
	})

	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}

	return keys, nil
}
