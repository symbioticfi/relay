package entity_processor

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto/blsBn254"
)

// Validators are looked up by the on-chain G1 key, so a gossiped signature that
// keeps a victim's G1 but carries an attacker's G2 must not verify. Otherwise
// any peer could sign under any validator's identity, and the per-index dedup
// would then censor that validator's genuine signature.
func TestEntityProcessor_ProcessSignature_RejectsUnboundG2(t *testing.T) {
	t.Parallel()

	for name, newRepo := range backends() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t)
			epoch := symbiotic.Epoch(700)
			req := randomSignatureRequest(t, epoch)

			_, privateKeys := setupValidatorSetHeader(t, repo, epoch, big.NewInt(1000))

			processor, err := NewEntityProcessor(Config{
				Repo:                     repo,
				Aggregator:               createMockAggregator(t),
				AggProofSignal:           createMockAggProofSignal(t),
				SignatureProcessedSignal: createMockSignatureProcessedSignal(t),
				Metrics:                  doNothingMetrics{},
			})
			require.NoError(t, err)

			// The attacker knows only the victim's public G1 identity.
			victimPub := privateKeys[0][req.KeyTag].PublicKey()
			victimG1Compressed := victimPub.Raw()[:32]

			messageHash, err := crypto.HashMessage(req.KeyTag.Type(), req.Message)
			require.NoError(t, err)

			// Sign with an arbitrary scalar and publish the matching G2.
			attackerScalar := big.NewInt(0xDEADBEEF)
			_, _, _, g2Gen := bn254.Generators()
			var attackerG2 bn254.G2Affine
			attackerG2.ScalarMultiplication(&g2Gen, attackerScalar)

			g1Hash, err := blsBn254.HashToG1(messageHash)
			require.NoError(t, err)
			var sig bn254.G1Affine
			sig.ScalarMultiplication(g1Hash, attackerScalar)

			attackerG2Compressed := attackerG2.Bytes()
			forgedRaw := append(append([]byte{}, victimG1Compressed...), attackerG2Compressed[:]...)
			forgedPub, err := crypto.NewPublicKey(symbiotic.KeyTypeBlsBn254, forgedRaw)
			require.NoError(t, err)
			require.Equal(t, victimPub.OnChain(), forgedPub.OnChain(),
				"forgery resolves to the victim's on-chain identity")

			forged := symbiotic.Signature{
				KeyTag:      req.KeyTag,
				Epoch:       epoch,
				MessageHash: messageHash,
				Signature:   sig.Marshal(),
				PublicKey:   forgedPub,
			}

			require.Error(t, processor.ProcessSignature(t.Context(), forged, false),
				"forged signature must not be accepted under the victim's identity")

			_, err = repo.GetSignatureMap(t.Context(), forged.RequestID())
			require.Error(t, err, "nothing recorded for the victim")

			// The victim's genuine signature is still accepted afterwards.
			genuine := signatureExtendedForRequest(t, privateKeys[0][req.KeyTag], req)
			require.NoError(t, processor.ProcessSignature(t.Context(), genuine, false))
		})
	}
}
