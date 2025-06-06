package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/network"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
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
		return fmt.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	entityMessage := entity.P2PSignatureHashMessage{
		Message: entity.SignatureHashMessage{
			Request: entity.SignatureRequest{
				KeyTag:        entity.KeyTag(signatureGenerated.Request.KeyTag),
				RequiredEpoch: signatureGenerated.Request.RequiredEpoch,
				Message:       signatureGenerated.Request.MessageHash,
			},
			Signature: signatureGenerated.Signature,
			PublicKey: signatureGenerated.PublicKey,
			HashType:  entity.HashType(signatureGenerated.HashType),
		},
		Info: entity.SenderInfo{
			Type:      info.Type,
			Sender:    info.Sender,
			Timestamp: info.Timestamp,
		},
	}

	err = s.signatureHashHandler(ctx, entityMessage)
	if err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}

	return nil
}

func (s *Service) handleStreamAggregatedProof(ctx context.Context, stream network.Stream) error {
	var signaturesAggregated signaturesAggregatedDTO
	info, err := unmarshalMessage(stream, &signaturesAggregated)
	if err != nil {
		return fmt.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	g1, err := bls.DeserializeG1(signaturesAggregated.PublicKeyG1)
	if err != nil {
		return fmt.Errorf("failed to deserialize G1 public key: %w", err)
	}
	entityMessage := entity.P2PSignaturesAggregatedMessage{
		Message: entity.SignaturesAggregatedMessage{
			PublicKeyG1: g1,
			Request: entity.SignatureRequest{
				KeyTag:        entity.KeyTag(signaturesAggregated.Request.KeyTag),
				RequiredEpoch: signaturesAggregated.Request.RequiredEpoch,
				Message:       signaturesAggregated.Request.MessageHash,
			},
			Proof: entity.AggregationProof{
				VerificationType: entity.VerificationType(signaturesAggregated.Proof.VerificationType),
				MessageHash:      signaturesAggregated.Proof.MessageHash,
				Proof:            signaturesAggregated.Proof.Proof,
			},
			HashType: entity.HashType(signaturesAggregated.HashType),
		},
		Info: entity.SenderInfo{
			Type:      info.Type,
			Sender:    info.Sender,
			Timestamp: info.Timestamp,
		},
	}

	err = s.signaturesAggregatedHandler(ctx, entityMessage)
	if err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}

	return nil
}

func unmarshalMessage(stream network.Stream, v interface{}) (entity.SenderInfo, error) {
	defer stream.Close()

	data := make([]byte, 1024*1024) // 1MB buffer
	n, err := stream.Read(data)
	if err != nil {
		return entity.SenderInfo{}, fmt.Errorf("failed to read from stream: %w", err)
	}

	var message p2pMessage
	if err := json.Unmarshal(data[:n], &message); err != nil {
		return entity.SenderInfo{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if err := json.Unmarshal(message.Data, v); err != nil {
		return entity.SenderInfo{}, fmt.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	return entity.SenderInfo{
		Type:      message.Type,
		Sender:    message.Sender,
		Timestamp: message.Timestamp,
	}, nil
}
