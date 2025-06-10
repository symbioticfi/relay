package p2p

import (
	"context"
	"encoding/json"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	dto := signaturesAggregatedDTO{
		RequestHash: msg.RequestHash,
		KeyTag:      uint8(msg.KeyTag),
		Epoch:       msg.Epoch,
		AggregationProof: aggregationProofDTO{
			MessageHash:      msg.AggregationProof.MessageHash,
			Proof:            msg.AggregationProof.Proof,
			VerificationType: uint32(msg.AggregationProof.VerificationType),
		},
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return errors.Errorf("failed to marshal signatures aggregated message: %w", err)
	}

	// send to ourselves first
	err = s.signaturesAggregatedHandler(ctx, entity.SenderInfo{
		Sender: "",
	}, msg)
	if err != nil {
		return errors.Errorf("failed to handle signatures aggregated message: %w", err)
	}

	return s.broadcast(ctx, messageTypeSignaturesAggregated, data)
}

type aggregationProofDTO struct {
	VerificationType uint32 `json:"verificationType"`
	MessageHash      []byte `json:"messageHash"`
	Proof            []byte `json:"proof"`
}
type signaturesAggregatedDTO struct {
	RequestHash      [32]byte            `json:"requestHash"`
	KeyTag           uint8               `json:"keyTag"`
	Epoch            uint64              `json:"epoch"`
	AggregationProof aggregationProofDTO `json:"proof"`
}
