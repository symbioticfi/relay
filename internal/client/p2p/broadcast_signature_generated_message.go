package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureExtended) error {
	dto := prototypes.SignatureGenerated{
		RequestId: msg.RequestID().Bytes(),
		KeyTag:    uint32(msg.KeyTag),
		Epoch:     uint64(msg.Epoch),
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
