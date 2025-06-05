package p2p

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureHashMessage) error {
	dto := signatureGeneratedDTO{
		MessageHash: msg.MessageHash,
		Signature:   msg.Signature,
		PublicKey:   msg.PublicKey,
		KeyTag:      uint8(msg.KeyTag),
		HashType:    string(msg.HashType),
		//ValsetHeaderTimestamp: msg.ValsetHeaderTimestamp, // todo ilya
		Epoch: msg.Epoch,
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

type signatureGeneratedDTO struct {
	// todo ilya add version
	MessageHash           []byte   `json:"message_hash"`
	Signature             []byte   `json:"signature"`
	PublicKey             []byte   `json:"public_key"`
	KeyTag                uint8    `json:"key_tag"`
	HashType              string   `json:"hash_type"`
	ValsetHeaderTimestamp *big.Int `json:"valset_header_timestamp"`
	Epoch                 *big.Int `json:"epoch"`
}
