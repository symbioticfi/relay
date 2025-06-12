package p2p

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/network"

	"middleware-offchain/core/entity"
	p2pEntity "middleware-offchain/internal/entity"
)

func handleStreamWrapper(ctx context.Context, f func(ctx context.Context, stream network.Stream) error) func(stream network.Stream) {
	return func(stream network.Stream) {
		if err := f(ctx, stream); err != nil {
			slog.ErrorContext(ctx, "Failed to handle stream", "error", err)
		}
	}
}

func (s *Service) handleStreamSignedHash(ctx context.Context, stream network.Stream) error {
	var signatureGenerated signatureGeneratedDTO
	err := unmarshalMessage(stream, &signatureGenerated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	msg := entity.SignatureMessage{
		RequestHash: signatureGenerated.RequestHash,
		KeyTag:      entity.KeyTag(signatureGenerated.KeyTag),
		Epoch:       signatureGenerated.Epoch,
		Signature: entity.Signature{
			PublicKey:   signatureGenerated.Signature.PublicKey,
			Signature:   signatureGenerated.Signature.Signature,
			MessageHash: signatureGenerated.Signature.MessageHash,
		},
	}

	si := p2pEntity.SenderInfo{
		Sender: stream.Conn().RemotePeer().String(),
	}

	s.signatureHashHandler.Emit(ctx, p2pEntity.P2PMessage[entity.SignatureMessage]{
		SenderInfo: si,
		Message:    msg,
	})

	return nil
}

func (s *Service) handleStreamAggregatedProof(ctx context.Context, stream network.Stream) error {
	var signaturesAggregated signaturesAggregatedDTO
	err := unmarshalMessage(stream, &signaturesAggregated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	msg := entity.AggregatedSignatureMessage{
		RequestHash: signaturesAggregated.RequestHash,
		KeyTag:      entity.KeyTag(signaturesAggregated.KeyTag),
		Epoch:       signaturesAggregated.Epoch,
		AggregationProof: entity.AggregationProof{
			VerificationType: entity.VerificationType(signaturesAggregated.AggregationProof.VerificationType),
			MessageHash:      signaturesAggregated.AggregationProof.MessageHash,
			Proof:            signaturesAggregated.AggregationProof.Proof,
		},
	}
	si := p2pEntity.SenderInfo{
		Sender: stream.Conn().RemotePeer().String(),
	}

	s.signaturesAggregatedHandler.Emit(ctx, p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]{
		SenderInfo: si,
		Message:    msg,
	})

	return nil
}

func unmarshalMessage(stream network.Stream, v interface{}) error {
	defer stream.Close()

	data := make([]byte, 1024*1024) // 1MB buffer
	n, err := stream.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.Errorf("failed to read from stream: %w", err)
	}

	var message p2pMessage
	if err := json.Unmarshal(data[:n], &message); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	if err := json.Unmarshal(message.Data, v); err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	return nil
}
