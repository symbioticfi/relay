//go:build manual

package eth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManual_GetEip712Domain(t *testing.T) {
	eth, err := NewEthClient(Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		MasterAddress:  "0xF91E4B4166AD3eafDE95FeB6402560FCAb881690",
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	domain, err := eth.GetEip712Domain(t.Context())
	require.NoError(t, err)
	fmt.Println(domain) // TODO remove
}

func TestManual_GetCurrentValsetTimestamp(t *testing.T) {
	eth, err := NewEthClient(Config{
		MasterRPCURL:   "http://127.0.0.1:8545",
		MasterAddress:  "0xF91E4B4166AD3eafDE95FeB6402560FCAb881690",
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	domain, err := eth.GetCurrentValsetTimestamp(t.Context())
	require.NoError(t, err)
	fmt.Println(domain) // TODO remove
}
