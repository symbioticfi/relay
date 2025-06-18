//go:build manual

package evm

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManual_GetEip712Domain(t *testing.T) {
	eth := initClient(t)

	domain, err := eth.GetEip712Domain(t.Context())
	require.NoError(t, err)
	fmt.Println(domain)
}

func initClient(t *testing.T) *Client {
	b, ok := new(big.Int).SetString("1000000000000000001", 10)
	require.True(t, ok)

	eth, err := NewEVMClient(Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		DriverAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
		RequestTimeout: time.Minute,
		PrivateKey:     b.FillBytes(make([]byte, 32)),
	})
	require.NoError(t, err)
	return eth
}
