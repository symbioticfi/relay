package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/symbioticfi/relay/core/entity"
	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

func (s *Service) BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	dto := prototypes.SignaturesAggregated{
		RequestHash: msg.RequestHash.Bytes(),
		KeyTag:      uint32(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		AggregationProof: &prototypes.AggregationProof{
			MessageHash:      msg.AggregationProof.MessageHash,
			Proof:            msg.AggregationProof.Proof,
			VerificationType: uint32(msg.AggregationProof.VerificationType),
		},
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signatures aggregated message: %w", err)
	}

	return s.broadcast(ctx, topicAggProofReady, data)
}
