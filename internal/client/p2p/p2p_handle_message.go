package p2p

import (
	"github.com/go-errors/errors"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) handleSignatureReadyMessage(pubSubMsg *pubsub.Message) error {
	var signatureGenerated prototypes.SignatureGenerated
	err := unmarshalMessage(pubSubMsg, &signatureGenerated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	// Validate the signatureGenerated message
	if len(signatureGenerated.GetRequestId()) > maxRequestIDSize {
		return errors.Errorf("request id size exceeds maximum allowed size: %d bytes", maxRequestIDSize)
	}
	if len(signatureGenerated.GetSignature().GetPublicKey()) > maxPubKeySize {
		return errors.Errorf("public key size exceeds maximum allowed size: %d bytes", maxPubKeySize)
	}
	if len(signatureGenerated.GetSignature().GetSignature()) > maxSignatureSize {
		return errors.Errorf("signature size exceeds maximum allowed size: %d bytes", maxSignatureSize)
	}
	if len(signatureGenerated.GetSignature().GetMessageHash()) > maxMsgHashSize {
		return errors.Errorf("message hash size exceeds maximum allowed size: %d bytes", maxMsgHashSize)
	}

	msg := entity.SignatureExtended{
		KeyTag:      entity.KeyTag(signatureGenerated.GetKeyTag()),
		Epoch:       entity.Epoch(signatureGenerated.GetEpoch()),
		PublicKey:   signatureGenerated.GetSignature().GetPublicKey(),
		Signature:   signatureGenerated.GetSignature().GetSignature(),
		MessageHash: signatureGenerated.GetSignature().GetMessageHash(),
	}

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	return s.signatureReceivedHandler.Emit(p2pEntity.P2PMessage[entity.SignatureExtended]{
		SenderInfo: si,
		Message:    msg,
	})
}

func (s *Service) handleAggregatedProofReadyMessage(pubSubMsg *pubsub.Message) error {
	var signaturesAggregated prototypes.SignaturesAggregated
	err := unmarshalMessage(pubSubMsg, &signaturesAggregated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signatureGenerated message: %w", err)
	}

	// Validate the signaturesAggregated message
	if len(signaturesAggregated.GetRequestId()) > maxRequestIDSize {
		return errors.Errorf("request id size exceeds maximum allowed size: %d bytes", maxRequestIDSize)
	}
	if len(signaturesAggregated.GetMessageHash()) > maxMsgHashSize {
		return errors.Errorf("aggregation proof message hash size exceeds maximum allowed size: %d bytes", maxMsgHashSize)
	}
	if len(signaturesAggregated.GetAggregationProof().GetProof()) > maxProofSize {
		return errors.Errorf("aggregation proof size exceeds maximum allowed size: %d bytes", maxProofSize)
	}

	msg := entity.AggregationProof{
		KeyTag:      entity.KeyTag(signaturesAggregated.GetKeyTag()),
		Epoch:       entity.Epoch(signaturesAggregated.GetEpoch()),
		MessageHash: signaturesAggregated.GetMessageHash(),
		Proof:       signaturesAggregated.GetAggregationProof().GetProof(),
	}

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	return s.signaturesAggregatedHandler.Emit(p2pEntity.P2PMessage[entity.AggregationProof]{
		SenderInfo: si,
		Message:    msg,
	})
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

func unmarshalMessage(msg *pubsub.Message, v proto.Message) error {
	var message prototypes.P2PMessage
	if err := proto.Unmarshal(msg.GetData(), &message); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	if err := proto.Unmarshal(message.GetData(), v); err != nil {
		return errors.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}
