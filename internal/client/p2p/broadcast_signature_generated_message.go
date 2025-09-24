package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/symbioticfi/relay/core/entity"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureExtended) error {
	dto := prototypes.SignatureGenerated{
		SignatureTargetId: msg.SignatureTargetID().Bytes(),
		KeyTag:            uint32(msg.KeyTag),
		Epoch:             uint64(msg.Epoch),
		Signature: &prototypes.Signature{
			MessageHash: msg.MessageHash,
			PublicKey:   msg.PublicKey,
			Signature:   msg.Signature,
		},
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	return s.broadcast(ctx, topicSignatureReady, data)
}
