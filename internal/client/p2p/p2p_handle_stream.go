package p2p

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p/core/network"

	"middleware-offchain/internal/entity"
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
	info, err := unmarshalMessage(stream, &signatureGenerated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	entityMessage := entity.P2PSignatureHashMessage{
		Message: entity.SignatureMessage{
			RequestHash: signatureGenerated.RequestHash,
			KeyTag:      entity.KeyTag(signatureGenerated.KeyTag),
			Epoch:       signatureGenerated.Epoch,
			Signature: entity.Signature{
				PublicKey:   signatureGenerated.Signature.PublicKey,
				Signature:   signatureGenerated.Signature.Signature,
				MessageHash: signatureGenerated.Signature.MessageHash,
			},
		},
		Info: entity.SenderInfo{
			Type:      info.Type,
			Sender:    info.Sender,
			Timestamp: info.Timestamp,
		},
	}

	err = s.signatureHashHandler(ctx, entityMessage)
	if err != nil {
		return errors.Errorf("failed to handle message: %w", err)
	}

	return nil
}

func (s *Service) handleStreamAggregatedProof(ctx context.Context, stream network.Stream) error {
	var signaturesAggregated signaturesAggregatedDTO
	info, err := unmarshalMessage(stream, &signaturesAggregated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	entityMessage := entity.P2PSignaturesAggregatedMessage{
		Message: entity.AggregatedSignatureMessage{
			RequestHash: signaturesAggregated.RequestHash,
			KeyTag:      entity.KeyTag(signaturesAggregated.KeyTag),
			Epoch:       signaturesAggregated.Epoch,
			AggregationProof: entity.AggregationProof{
				VerificationType: entity.VerificationType(signaturesAggregated.AggregationProof.VerificationType),
				MessageHash:      signaturesAggregated.AggregationProof.MessageHash,
				Proof:            signaturesAggregated.AggregationProof.Proof,
			},
		},
		Info: entity.SenderInfo{
			Type:      info.Type,
			Sender:    info.Sender,
			Timestamp: info.Timestamp,
		},
	}

	err = s.signaturesAggregatedHandler(ctx, entityMessage)
	if err != nil {
		return errors.Errorf("failed to handle message: %w", err)
	}

	return nil
}

func unmarshalMessage(stream network.Stream, v interface{}) (entity.SenderInfo, error) {
	defer stream.Close()

	data := make([]byte, 1024*1024) // 1MB buffer
	n, err := stream.Read(data)
	if err != nil {
		return entity.SenderInfo{}, errors.Errorf("failed to read from stream: %w", err)
	}

	var message p2pMessage
	if err := json.Unmarshal(data[:n], &message); err != nil {
		return entity.SenderInfo{}, errors.Errorf("failed to unmarshal message: %w", err)
	}

	if err := json.Unmarshal(message.Data, v); err != nil {
		return entity.SenderInfo{}, errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	return entity.SenderInfo{
		Type:      message.Type,
		Sender:    message.Sender,
		Timestamp: message.Timestamp,
	}, nil
}
