package signer_app

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

func (s *SignerApp) signMessage(ctx context.Context, message []byte) error {
	messageHash := crypto.Keccak256(message)
	headerSignature, err := s.cfg.KeyPair.Sign(messageHash)
	if err != nil {
		return errors.Errorf("failed to sign message hash: %w", err)
	}

	slog.DebugContext(ctx, "message hash signed, sending via p2p", "headerSignature", headerSignature)

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: messageHash,
		KeyTag:      15, // todo ilya: pass from config or from another place
		Signature:   bls.SerializeG1(headerSignature),
		PublicKeyG1: bls.SerializeG1(&s.cfg.KeyPair.PublicKeyG1),
		PublicKeyG2: bls.SerializeG2(&s.cfg.KeyPair.PublicKeyG2),
		HashType:    entity.HashTypeMessage,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signed hash message: %w", err)
	}

	slog.DebugContext(ctx, "message hash sent p2p")

	return nil
}
