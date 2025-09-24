package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/symbioticfi/relay/core/entity"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregationProof) error {
	dto := prototypes.SignaturesAggregated{
		SignatureTargetId: msg.SignatureTargetID().Bytes(),
		KeyTag:            uint32(msg.KeyTag),
		Epoch:             uint64(msg.Epoch),
		AggregationProof: &prototypes.AggregationProof{
			MessageHash: msg.MessageHash,
			Proof:       msg.Proof,
		},
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signatures aggregated message: %w", err)
	}

	return s.broadcast(ctx, topicAggProofReady, data)
}
