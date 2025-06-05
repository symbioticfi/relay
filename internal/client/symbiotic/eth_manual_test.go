//go:build manual

package symbiotic

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManual_GetEip712Domain(t *testing.T) {
	eth, err := NewEVMClient(Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		MasterAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	domain, err := eth.GetEip712Domain(t.Context())
	require.NoError(t, err)
	fmt.Println(domain) // TODO remove
}

func TestManual_GetCurrentValsetTimestamp(t *testing.T) {
	eth, err := NewEVMClient(Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		MasterAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	domain, err := eth.GetCurrentValsetTimestamp(t.Context())
	require.NoError(t, err)
	fmt.Println(domain) // TODO remove
}
