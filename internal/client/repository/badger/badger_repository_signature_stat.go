package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
)

func keySignatureStat(reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("signature_stat:%s", reqHash.Hex()))
}

func signatureStatToBytes(stat entity.SignatureStat) ([]byte, error) {
	return json.Marshal(signatureStatDTO{
		ReqHash: stat.ReqHash,
		StatMap: lo.MapKeys(stat.StatMap, func(v time.Time, k entity.SignatureStatStage) string {
			switch k {
			case entity.SignatureStatStageSignRequestReceived:
				return "sign_request_received"
			case entity.SignatureStatStageSignCompleted:
				return "sign_completed"
			case entity.SignatureStatStageAggQuorumReached:
				return "agg_quorum_reached"
			case entity.SignatureStatStageAggCompleted:
				return "agg_completed"
			case entity.SignatureStatStageAggProofReceived:
				return "agg_proof_received"
			case entity.SignatureStatStageUnknown:
				fallthrough
			default:
				return "unknown"
			}
		}),
	})
}

func bytesToSignatureStat(b []byte) (entity.SignatureStat, error) {
	var dto signatureStatDTO
	if err := json.Unmarshal(b, &dto); err != nil {
		return entity.SignatureStat{}, errors.Errorf("failed to unmarshal signature stat: %w", err)
	}

	return entity.SignatureStat{
		ReqHash: dto.ReqHash,
		StatMap: lo.MapKeys(dto.StatMap, func(v time.Time, k string) entity.SignatureStatStage {
			switch k {
			case "sign_request_received":
				return entity.SignatureStatStageSignRequestReceived
			case "sign_completed":
				return entity.SignatureStatStageSignCompleted
			case "agg_quorum_reached":
				return entity.SignatureStatStageAggQuorumReached
			case "agg_completed":
				return entity.SignatureStatStageAggCompleted
			case "agg_proof_received":
				return entity.SignatureStatStageAggProofReceived
			default:
				return entity.SignatureStatStageUnknown
			}
		}),
	}, nil
}

func (r *Repository) UpdateSignatureStat(_ context.Context, reqHash common.Hash, s entity.SignatureStatStage, t time.Time) (entity.SignatureStat, error) {
	var oldStat entity.SignatureStat
	return oldStat, r.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(keySignatureStat(reqHash))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get signature stat: %w", err)
		}
		if errors.Is(err, badger.ErrKeyNotFound) {
			oldStat = entity.SignatureStat{
				ReqHash: reqHash,
				StatMap: make(map[entity.SignatureStatStage]time.Time),
			}
		} else {
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature stat value: %w", err)
			}

			oldStat, err = bytesToSignatureStat(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature stat: %w", err)
			}
		}
		oldStat.StatMap[s] = t

		bytes, err := signatureStatToBytes(oldStat)
		if err != nil {
			return errors.Errorf("failed to marshal signature stat: %w", err)
		}

		err = txn.Set(keySignatureStat(reqHash), bytes)
		if err != nil {
			return errors.Errorf("failed to store signature stat: %w", err)
		}
		return nil
	})
}

type signatureStatDTO struct {
	ReqHash common.Hash          `json:"req_hash"`
	StatMap map[string]time.Time `json:"stat_map"`
}
