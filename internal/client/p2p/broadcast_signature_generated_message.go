package p2p

import (
	"context"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) BroadcastSignatureGeneratedMessage(ctx context.Context, msg symbiotic.Signature) error {
	dto := prototypes.Signature{
		KeyTag:      uint32(msg.KeyTag),
		Epoch:       uint64(msg.Epoch),
		MessageHash: msg.MessageHash,
		PublicKey:   msg.PublicKey.Raw(),
		Signature:   msg.Signature,
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		carrier := propagation.MapCarrier{}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		dto.TraceContext = carrier
	}

	data, err := proto.Marshal(&dto)
	if err != nil {
		return errors.Errorf("failed to marshal signature generated message: %w", err)
	}

	return s.broadcast(ctx, topicSignatureReady, data)
}
