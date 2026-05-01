package bbolt

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func randomRequestID(t *testing.T) common.Hash {
	t.Helper()
	req := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: symbiotic.Epoch(randomBigInt(t).Uint64()),
		Message:       randomBytes(t, 32),
	}
	priv, err := crypto.GeneratePrivateKey(req.KeyTag.Type())
	require.NoError(t, err)
	_, messageHash, err := priv.Sign(req.Message)
	require.NoError(t, err)

	sig := symbiotic.Signature{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		PublicKey:   priv.PublicKey(),
	}
	return sig.RequestID()
}

func TestRepository_GetSignatureMap_NotFound(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	nonExistentHash := randomRequestID(t)
	_, err := repo.GetSignatureMap(context.Background(), nonExistentHash)
	require.Error(t, err)
	assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
}
