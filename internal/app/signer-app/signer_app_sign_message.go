package signer_app

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *SignerApp) signMessage(ctx context.Context, message []byte, keyTag entity.KeyTag) error {
	messageHash, err := s.cfg.Signer.Hash(keyTag, message)
	if err != nil {
		return errors.Errorf("failed to hash message: %w", err)
	}

	slog.InfoContext(ctx, "got request to sign message",
		"originalMessage", hex.EncodeToString(message),
		"messageHash", hex.EncodeToString(messageHash),
	)

	messageSignature, err := s.cfg.Signer.Sign(keyTag, messageHash)
	if err != nil {
		return errors.Errorf("failed to sign message hash: %w", err)
	}

	slog.DebugContext(ctx, "message hash signed, sending via p2p", "headerSignature", messageSignature)

	// TODO ilya
	//epoch, err := s.cfg.EthClient.GetCurrentValsetEpoch(ctx)
	//if err != nil {
	//	return errors.Errorf("failed to get current epoch: %w", err)
	//}

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: messageHash,
		KeyTag:      keyTag,
		Signature:   messageSignature.Signature,
		PublicKey:   messageSignature.PublicKey,
		HashType:    entity.HashTypeMessage,
		//Epoch:                 epoch,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signed hash message: %w", err)
	}

	slog.DebugContext(ctx, "message hash sent p2p")

	return nil
}
