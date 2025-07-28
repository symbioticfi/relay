package p2p

import (
	"context"
	"encoding/json"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// todo ilya, how to name signature generated or signature ready?
func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error {
	dto := signatureGeneratedDTO{
		RequestHash: msg.RequestHash,
		KeyTag:      uint8(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		Signature: signatureDTO{
			MessageHash: msg.Signature.MessageHash,
			PublicKey:   msg.Signature.PublicKey,
			Signature:   msg.Signature.Signature,
		},
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	return s.broadcast(ctx, topicSignatureReady, data)
}

type signatureDTO struct {
	MessageHash []byte `json:"messageHash"`
	Signature   []byte `json:"signature"`
	PublicKey   []byte `json:"publicKey"`
}
type signatureGeneratedDTO struct {
	RequestHash [32]byte     `json:"requestHash"`
	KeyTag      uint8        `json:"keyTag"`
	Epoch       uint64       `json:"epoch"`
	Signature   signatureDTO `json:"signature"`
}
