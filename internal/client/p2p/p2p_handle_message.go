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

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	s.signatureReceivedHandler.Emit(ctx, p2pEntity.P2PMessage[entity.SignatureMessage]{
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

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	s.signaturesAggregatedHandler.Emit(ctx, p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]{
		SenderInfo: si,
		Message:    msg,
	})

	return nil
}

func extractSenderInfo(pubSubMsg *pubsub.Message) (p2pEntity.SenderInfo, error) {
	// try to extract public key from sender peer.ID
	pubKey, err := pubSubMsg.ReceivedFrom.ExtractPublicKey()
	if err != nil {
		return p2pEntity.SenderInfo{}, errors.Errorf("failed to extract public key from received message from peer %s: %w", pubSubMsg.ReceivedFrom.String(), err)
	}

	raw, err := pubKey.Raw()
	if err != nil {
		return p2pEntity.SenderInfo{}, errors.Errorf("failed to get raw public key from peer %s: %w", pubSubMsg.ReceivedFrom.String(), err)
	}

	return p2pEntity.SenderInfo{
		Sender:    pubSubMsg.ReceivedFrom.String(),
		PublicKey: raw,
	}, nil
}

func unmarshalMessage(msg *pubsub.Message, v interface{}) error {
	var message p2pMessage
	if err := json.Unmarshal(msg.GetData(), &message); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	if err := json.Unmarshal(message.Data, v); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}
