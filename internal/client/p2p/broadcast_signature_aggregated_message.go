package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg symbiotic.AggregationProof) error {
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

	return s.broadcast(ctx, topicAggProofReady, data)
}
