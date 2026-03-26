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

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg symbiotic.AggregationProof) error {
	start := time.Now()
	dto := prototypes.AggregationProof{
		KeyTag:      uint32(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		MessageHash: msg.MessageHash,
		Proof:       msg.Proof,
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signatures aggregated message: %w", err)
	}

	err = s.broadcast(ctx, topicAggProofReady, data)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "Broadcasted signatures aggregated message", "duration", time.Since(start))

	return nil
}
