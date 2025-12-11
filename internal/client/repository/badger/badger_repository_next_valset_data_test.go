package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestRepository_SaveNextValsetData_HappyPath(t *testing.T) {
	repo := setupTestRepository(t)

	data := newTestNextValsetData(t)
	requestID := data.ValidatorSetMetadata.RequestID

	// Save the data
	err := repo.SaveNextValsetData(t.Context(), data)
	require.NoError(t, err, "SaveNextValsetData should succeed")

	// Verify data was saved by reading it back
	// Verify previous validator set was saved
	savedPrevVS, err := repo.GetValidatorSetByEpoch(t.Context(), data.PrevValidatorSet.Epoch)
	require.NoError(t, err)
	require.Equal(t, data.PrevValidatorSet, savedPrevVS)

	// Verify next validator set was saved
	savedNextVS, err := repo.GetValidatorSetByEpoch(t.Context(), data.NextValidatorSet.Epoch)
	require.NoError(t, err)
	require.Equal(t, data.NextValidatorSet, savedNextVS)

	// Verify previous config was saved
	savedPrevConfig, err := repo.GetConfigByEpoch(t.Context(), data.PrevValidatorSet.Epoch)
	require.NoError(t, err)
	require.Equal(t, data.PrevNetworkConfig, savedPrevConfig)

	// Verify next config was saved
	savedNextConfig, err := repo.GetConfigByEpoch(t.Context(), data.NextValidatorSet.Epoch)
	require.NoError(t, err)
	require.Equal(t, data.NextNetworkConfig, savedNextConfig)

	// Verify signature request was saved
	savedSigReq, err := repo.GetSignatureRequest(t.Context(), requestID)
	require.NoError(t, err)
	require.Equal(t, *data.SignatureRequest, savedSigReq)
}

func TestRepository_SaveNextValsetData_IgnoresExistingMappings(t *testing.T) {
	repo := setupTestRepository(t)

	data := newTestNextValsetData(t)
	err := repo.SaveNextValsetData(t.Context(), data)
	require.NoError(t, err)

	// Remove pending entry so the second save only hits the mapping duplication path.
	err = repo.removeProofCommitPending(t.Context(), data.NextValidatorSet.Epoch)
	require.NoError(t, err)

	err = repo.SaveNextValsetData(t.Context(), data)
	require.NoError(t, err)
}

func newTestNextValsetData(t *testing.T) entity.NextValsetData {
	t.Helper()

	prevValidatorSet := randomValidatorSet(t, 1)
	prevValidatorSet.Status = symbiotic.HeaderCommitted
	prevNetworkConfig := randomNetworkConfig(t)

	nextValidatorSet := randomValidatorSet(t, 2)
	nextValidatorSet.Status = symbiotic.HeaderDerived
	nextNetworkConfig := randomNetworkConfig(t)

	requestID := common.BytesToHash(randomBytes(t, 32))
	signatureRequest := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: nextValidatorSet.Epoch,
		Message:       randomBytes(t, 100),
	}

	metadata := symbiotic.ValidatorSetMetadata{
		RequestID:      requestID,
		Epoch:          nextValidatorSet.Epoch,
		ExtraData:      []symbiotic.ExtraData{},
		CommitmentData: randomBytes(t, 64),
	}

	return entity.NextValsetData{
		PrevValidatorSet:     prevValidatorSet,
		PrevNetworkConfig:    prevNetworkConfig,
		NextValidatorSet:     nextValidatorSet,
		NextNetworkConfig:    nextNetworkConfig,
		SignatureRequest:     &signatureRequest,
		ValidatorSetMetadata: metadata,
	}
}
