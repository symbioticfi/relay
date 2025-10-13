package p2p

import (
	"github.com/go-errors/errors"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"google.golang.org/protobuf/proto"

	prototypes "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *Service) handleSignatureReadyMessage(pubSubMsg *pubsub.Message) error {
	var signature prototypes.Signature
	err := unmarshalMessage(pubSubMsg, &signature)
	if err != nil {
		return errors.Errorf("failed to unmarshal signature message: %w", err)
	}

	// Validate the signature message
	if len(signature.GetPublicKey()) > maxPubKeySize {
		return errors.Errorf("public key size exceeds maximum allowed size: %d bytes", maxPubKeySize)
	}
	if len(signature.GetSignature()) > maxSignatureSize {
		return errors.Errorf("signature size exceeds maximum allowed size: %d bytes", maxSignatureSize)
	}
	if len(signature.GetMessageHash()) > maxMsgHashSize {
		return errors.Errorf("message hash size exceeds maximum allowed size: %d bytes", maxMsgHashSize)
	}

	pubKey, err := crypto.NewPublicKey(symbiotic.KeyTag(signature.GetKeyTag()).Type(), signature.GetPublicKey())
	if err != nil {
		return errors.Errorf("failed to parse public key: %w", err)
	}

	msg := symbiotic.Signature{
		KeyTag:      symbiotic.KeyTag(signature.GetKeyTag()),
		Epoch:       symbiotic.Epoch(signature.GetEpoch()),
		PublicKey:   pubKey,
		Signature:   signature.GetSignature(),
		MessageHash: signature.GetMessageHash(),
	}

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	return s.signatureReceivedHandler.Emit(p2pEntity.P2PMessage[symbiotic.Signature]{
		SenderInfo: si,
		Message:    msg,
	})
}

func (s *Service) handleAggregatedProofReadyMessage(pubSubMsg *pubsub.Message) error {
	var signaturesAggregated prototypes.AggregationProof
	err := unmarshalMessage(pubSubMsg, &signaturesAggregated)
	if err != nil {
		return errors.Errorf("failed to unmarshal signature message: %w", err)
	}

	// Validate the signaturesAggregated message
	if len(signaturesAggregated.GetMessageHash()) > maxMsgHashSize {
		return errors.Errorf("aggregation proof message hash size exceeds maximum allowed size: %d bytes", maxMsgHashSize)
	}
	if len(signaturesAggregated.GetProof()) > maxProofSize {
		return errors.Errorf("aggregation proof size exceeds maximum allowed size: %d bytes", maxProofSize)
	}

	msg := symbiotic.AggregationProof{
		KeyTag:      symbiotic.KeyTag(signaturesAggregated.GetKeyTag()),
		Epoch:       symbiotic.Epoch(signaturesAggregated.GetEpoch()),
		MessageHash: signaturesAggregated.GetMessageHash(),
		Proof:       signaturesAggregated.GetProof(),
	}

	si, err := extractSenderInfo(pubSubMsg)
	if err != nil {
		return errors.Errorf("failed to extract sender info from received message: %w", err)
	}

	return s.signaturesAggregatedHandler.Emit(p2pEntity.P2PMessage[symbiotic.AggregationProof]{
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
