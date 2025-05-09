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
		PublicKeyG1: msg.PublicKeyG1,
		PublicKeyG2: msg.PublicKeyG2,
		KeyTag:      msg.KeyTag,
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
	PublicKeyG1 []byte `json:"public_key_g1"`
	PublicKeyG2 []byte `json:"public_key_g2"`
	KeyTag      uint64 `json:"key_tag"`
}
