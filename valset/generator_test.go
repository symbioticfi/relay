package valset

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
)

func TestGenerator(t *testing.T) {
	// Define the large number
	privateKeyInt := new(big.Int)
	privateKeyInt.SetString("87191036493798670866484781455694320176667203290824056510541300741498740913410", 10)

	// Convert to bytes
	privateKeyBytes := privateKeyInt.Bytes()

	client, err := eth.NewEthClient("http://127.0.0.1:8545", "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f", privateKeyBytes)
	require.NoError(t, err)

	deriver, err := NewValsetDeriver(client)
	require.NoError(t, err)

	generator, err := NewValsetGenerator(deriver, client)
	require.NoError(t, err)

	header, err := generator.GenerateValidatorSetHeader(context.Background())
	require.NoError(t, err)

	fmt.Println(header)
}
