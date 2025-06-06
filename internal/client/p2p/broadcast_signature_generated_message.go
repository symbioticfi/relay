package p2p

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureHashMessage) error {
	dto := signatureGeneratedDTO{
		Request: signatureRequestDTO{
			KeyTag:        uint8(msg.Request.KeyTag),
			RequiredEpoch: msg.Request.RequiredEpoch,
			MessageHash:   msg.Request.Message,
		},
		Signature: msg.Signature,
		PublicKey: msg.PublicKey,
		HashType:  string(msg.HashType),
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	// send to ourselves first
	err = s.signatureHashHandler(ctx, entity.P2PSignatureHashMessage{
		Message: msg,
		Info: entity.SenderInfo{
			Type:      entity.P2PMessageTypeSignatureHash,
			Sender:    "",
			Timestamp: time.Now().Unix(),
		},
	})
	if err != nil {
		return errors.Errorf("failed to handle signature generated message: %w", err)
	}

	return s.broadcast(ctx, entity.P2PMessageTypeSignatureHash, data)
}

type signatureRequestDTO struct {
	KeyTag        uint8  `json:"key_tag"`
	RequiredEpoch uint64 `json:"required_epoch"`
	MessageHash   []byte `json:"message_hash"`
}
type signatureGeneratedDTO struct {
	Request               signatureRequestDTO `json:"request"`
	Signature             []byte              `json:"signature"`
	PublicKey             []byte              `json:"public_key"`
	HashType              string              `json:"hash_type"`
	ValsetHeaderTimestamp uint64              `json:"valset_header_timestamp"`
}
