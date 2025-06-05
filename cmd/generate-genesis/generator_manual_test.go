//go:build manual

package main

//
//import (
//	"context"
//	"fmt"
//	"log/slog"
//	"math/big"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/require"
//
//	"middleware-offchain/internal/client/symbiotic"
//	"middleware-offchain/internal/client/valset"
//)
//
//func TestManual_GenerateValidatorSetHeader(t *testing.T) {
//	// Define the large number
//	privateKeyInt := new(big.Int)
//	privateKeyInt.SetString("87191036493798670866484781455694320176667203290824056510541300741498740913410", 10)
//
//	client, err := symbiotic.NewEthClient(symbiotic.Config{
//		MasterRPCURL:   "http://127.0.0.1:8545",
//		MasterAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
//		PrivateKey:     privateKeyInt.Bytes(),
//		RequestTimeout: time.Minute,
//	})
//	require.NoError(t, err)
//
//	deriver, err := valset.NewDeriver(client)
//	require.NoError(t, err)
//
//	generator, err := valset.NewGenerator(deriver, client)
//	require.NoError(t, err)
//
//	header, err := generator.GenerateValidatorSetHeaderOnCapture(context.Background())
//	require.NoError(t, err)
//
//	jsonData, err := valsetHeaderMarshalJSON(header)
//	if err != nil {
//		t.Fatalf("Failed to marshal header to JSON: %v", err)
//	}
//	slog.Debug("Generated validator set header", "json", string(jsonData))
//
//	fmt.Println("Header:", string(jsonData))
//}
