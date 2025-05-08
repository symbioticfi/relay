package p2p

import (
	"context"
	"encoding/json"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error {
	dto := signatureGeneratedDTO{
		MessageHash: msg.MessageHash,
		Signature:   msg.Signature,
		PublicKey:   msg.PublicKey,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	return s.broadcast(ctx, entity.SignatureGenerated, data)
}

type signatureGeneratedDTO struct {
	MessageHash []byte `json:"message_hash"`
	Signature   []byte `json:"signature"`
	PublicKey   []byte `json:"public_key"`
}
