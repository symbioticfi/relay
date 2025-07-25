package p2p

import (
	"context"
	"encoding/json"

	"github.com/go-errors/errors"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
)

func (s *Service) handleSignatureReadyMessage(ctx context.Context, pubSubMsg *pubsub.Message) error {
	var signatureGenerated signatureGeneratedDTO
	err := unmarshalMessage(pubSubMsg, &signatureGenerated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	msg := entity.SignatureMessage{
		RequestHash: signatureGenerated.RequestHash,
		KeyTag:      entity.KeyTag(signatureGenerated.KeyTag),
		Epoch:       entity.Epoch(signatureGenerated.Epoch),
		Signature: entity.SignatureExtended{
			PublicKey:   signatureGenerated.Signature.PublicKey,
			Signature:   signatureGenerated.Signature.Signature,
			MessageHash: signatureGenerated.Signature.MessageHash,
		},
	}

	si := p2pEntity.SenderInfo{
		Sender: pubSubMsg.ReceivedFrom.String(),
	}

	s.signatureHashHandler.Emit(ctx, p2pEntity.P2PMessage[entity.SignatureMessage]{
		SenderInfo: si,
		Message:    msg,
	})

	return nil
}

func (s *Service) handleAggregatedProofReadyMessage(ctx context.Context, pubSubMsg *pubsub.Message) error {
	var signaturesAggregated signaturesAggregatedDTO
	err := unmarshalMessage(pubSubMsg, &signaturesAggregated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	msg := entity.AggregatedSignatureMessage{
		RequestHash: signaturesAggregated.RequestHash,
		KeyTag:      entity.KeyTag(signaturesAggregated.KeyTag),
		Epoch:       entity.Epoch(signaturesAggregated.Epoch),
		AggregationProof: entity.AggregationProof{
			VerificationType: entity.VerificationType(signaturesAggregated.AggregationProof.VerificationType),
			MessageHash:      signaturesAggregated.AggregationProof.MessageHash,
			Proof:            signaturesAggregated.AggregationProof.Proof,
		},
	}
	si := p2pEntity.SenderInfo{
		Sender: pubSubMsg.ReceivedFrom.String(),
	}

	s.signaturesAggregatedHandler.Emit(ctx, p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]{
		SenderInfo: si,
		Message:    msg,
	})

	return nil
}

func unmarshalMessage(msg *pubsub.Message, v interface{}) error {
	var message p2pMessage
	if err := json.Unmarshal(msg.GetData(), &message); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	if err := json.Unmarshal(message.Data, v); err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	return nil
}
