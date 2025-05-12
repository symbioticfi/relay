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

func TestGenerator(t *testing.T) {
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

	generator, err := NewValsetGenerator(deriver, client)
	require.NoError(t, err)

	header, err := generator.GenerateValidatorSetHeader(context.Background())
	require.NoError(t, err)

	jsonData, err := header.EncodeJSON()
	if err != nil {
		t.Fatalf("Failed to marshal header to JSON: %v", err)
	}
	slog.Debug("Generated validator set header", "json", string(jsonData))

	fmt.Println("Header:", string(jsonData))
}
