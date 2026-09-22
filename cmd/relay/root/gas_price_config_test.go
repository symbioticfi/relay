package root

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGasPriceFlagsRoundTrip(t *testing.T) {
	previousCmd, previousFile := rootCmd, configFile
	t.Cleanup(func() { rootCmd, configFile = previousCmd, previousFile })
	rootCmd = &cobra.Command{}
	rootCmd.SetContext(context.Background())
	addRootFlags(rootCmd)
	t.Setenv("SYMB_EVM_FALLBACK_GAS_PRICES", "")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, nil, 0600))
	// Repeated flags and comma-separated lists both pass through Viper's String/Set round trip.
	args := []string{"--config=" + configPath, "--driver.chain-id=1", "--driver.address=0x0000000000000000000000000000000000000001", "--api.listen=127.0.0.1:8080", "--p2p.listen=/ip4/127.0.0.1/tcp/0", "--evm.chains=http://127.0.0.1:8545", "--secret-keys=test/1/1/unused-test-value", "--evm.fallback-gas-prices=1=2,10=3", "--evm.fallback-gas-prices=1=4"}
	require.NoError(t, rootCmd.ParseFlags(args))
	require.NoError(t, initConfig(rootCmd, nil))
	require.Equal(t, CMDGasPriceMap{1: 4, 10: 3}, cfgFromCtx(rootCmd.Context()).Evm.FallbackGasPrices)
}

func TestGasPriceMapRejectsInvalidListWithoutMutation(t *testing.T) {
	for _, value := range []string{"1=4,broken", "1=4,invalid=3", "1=4,10=18446744073709551616"} {
		t.Run(value, func(t *testing.T) {
			prices := CMDGasPriceMap{1: 2}
			require.Error(t, prices.Set(value))
			require.Equal(t, CMDGasPriceMap{1: 2}, prices)
		})
	}
}
