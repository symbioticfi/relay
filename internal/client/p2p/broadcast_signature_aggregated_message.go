package p2p

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.SignaturesAggregatedMessage) error {
	dto := signaturesAggregatedDTO{
		PublicKeyG1: bls.SerializeG1(msg.PublicKeyG1),
		Proof:       msg.Proof,
		Message:     msg.Message,
		HashType:    string(msg.HashType),
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return errors.Errorf("failed to marshal signatures aggregated message: %w", err)
	}

	// send to ourselves first
	err = s.signaturesAggregatedHandler(ctx, entity.P2PSignaturesAggregatedMessage{
		Message: msg,
		Info: entity.SenderInfo{
			Type:      entity.P2PMessageTypeSignatureHash,
			Sender:    "",
			Timestamp: time.Now().Unix(),
		},
	})
	if err != nil {
		return errors.Errorf("failed to handle signatures aggregated message: %w", err)
	}

	return s.broadcast(ctx, entity.P2PMessageTypeSignaturesAggregated, data)
}

type signaturesAggregatedDTO struct {
	PublicKeyG1 []byte `json:"public_key_g1"`
	Proof       []byte `json:"proof"`
	Message     []byte `json:"message"`
	HashType    string `json:"hash_type"`
}
