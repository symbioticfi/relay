package p2p

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg symbiotic.Signature) error {
	start := time.Now()
	dto := prototypes.Signature{
		KeyTag:      uint32(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		MessageHash: msg.MessageHash,
		PublicKey:   msg.PublicKey.Raw(),
		Signature:   msg.Signature,
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	err = s.broadcast(ctx, topicSignatureReady, data)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Broadcasted signature generated message", "duration", time.Since(start))

	return nil
}
