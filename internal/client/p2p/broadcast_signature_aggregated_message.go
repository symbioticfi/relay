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
		Request: signatureRequestDTO{
			KeyTag:        uint8(msg.Request.KeyTag),
			RequiredEpoch: msg.Request.RequiredEpoch,
			MessageHash:   msg.Request.Message,
		},
		PublicKeyG1: bls.SerializeG1(msg.PublicKeyG1),
		Proof: aggregationProofDTO{
			VerificationType: uint32(msg.Proof.VerificationType),
			MessageHash:      msg.Proof.MessageHash,
			Proof:            msg.Proof.Proof,
		},
		HashType: string(msg.HashType),
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

type aggregationProofDTO struct {
	VerificationType uint32 `json:"verification_type"`
	MessageHash      []byte `json:"message_hash"`
	Proof            []byte `json:"proof"`
}
type signaturesAggregatedDTO struct {
	Request     signatureRequestDTO `json:"request"`
	PublicKeyG1 []byte              `json:"public_key_g1"`
	Proof       aggregationProofDTO `json:"proof"`
	HashType    string              `json:"hash_type"`
}
