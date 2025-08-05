package p2p

import (
	"context"

	"github.com/go-errors/errors"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"google.golang.org/protobuf/proto"

	"github.com/symbioticfi/relay/core/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error {
	dto := prototypes.SignatureGenerated{
		RequestHash: msg.RequestHash.Bytes(),
		KeyTag:      uint32(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		Signature: &prototypes.Signature{
			MessageHash: msg.Signature.MessageHash,
			PublicKey:   msg.Signature.PublicKey,
			Signature:   msg.Signature.Signature,
		},
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	return s.broadcast(ctx, topicSignatureReady, data)
}
