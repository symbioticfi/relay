package valset

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/samber/lo"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
)

func TestDeriver(t *testing.T) {
	// Define the large number
	privateKeyInt := new(big.Int)
	privateKeyInt.SetString("87191036493798670866484781455694320176667203290824056510541300741498740913410", 10)

	client, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     lo.ToPtr(privateKeyInt.Bytes()),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	deriver, err := NewValsetDeriver(client)
	require.NoError(t, err)

	timestamp, err := client.GetCaptureTimestamp(context.Background())
	require.NoError(t, err)
	slog.DebugContext(context.Background(), "Got capture timestamp", "timestamp", timestamp.String())

	validatorSet, err := deriver.GetValidatorSet(context.Background(), timestamp)
	require.NoError(t, err)

	validator, validatorRootTreeLocalIndex, validatorRootProof, err := validatorSet.ProveValidatorRoot(validatorSet.Validators[1].Operator)
	require.NoError(t, err)

	fmt.Printf("validatorRootProof: %x\n", validatorRootProof)
	fmt.Printf("validatorRootTreeLocalIndex: %d\n", validatorRootTreeLocalIndex)
	vault, vaultIndex, vaultRootProof, err := validator.ProveVaultRoot(validator.Vaults[1].Vault)
	require.NoError(t, err)

	fmt.Printf("vaultRootProof: %x\n", vaultRootProof)
	fmt.Printf("vaultIndex: %d\n", vaultIndex)

	vaultVotingPowerProof, err := vault.ProveVaultVotingPower()
	require.NoError(t, err)

	fmt.Printf("vaultVotingPowerProof: %x\n", vaultVotingPowerProof)
	fmt.Printf("vaultVotingPower: %x\n", vault.VotingPower)
}
