package network

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cmdhelpers "github.com/symbiotic/relay/internal/usecase/cmd-helpers"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewNetworkCmd() *cobra.Command {
	networkCmd.AddCommand(infoCmd)
	networkCmd.AddCommand(genesisCmd)

	initFlags()

	return networkCmd
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network tool",
}

type GlobalFlags struct {
	Chains        []string
	DriverAddress string
	DriverChainId uint64
	Epoch         uint64
}

type InfoFlags struct {
	Validators     bool
	ValidatorsFull bool
	Addresses      bool
	Settlement     bool
}

type GenesisFlags struct {
	Json    bool
	Commit  bool
	Output  string
	Secrets cmdhelpers.SecretKeyMapFlag
}

var globalFlags GlobalFlags
var infoFlags InfoFlags
var genesisFlags GenesisFlags

func initFlags() {
	networkCmd.PersistentFlags().StringSliceVarP(&globalFlags.Chains, "chains", "c", nil, "Chains rpc url, comma separated")
	networkCmd.PersistentFlags().StringVar(&globalFlags.DriverAddress, "driver-address", "", "Driver contract address")
	networkCmd.PersistentFlags().Uint64Var(&globalFlags.DriverChainId, "driver-chainid", 0, "Driver contract chain id")
	networkCmd.PersistentFlags().Uint64VarP(&globalFlags.Epoch, "epoch", "e", 0, "Network epoch to fetch info")
	if err := networkCmd.MarkPersistentFlagRequired("chains"); err != nil {
		panic(err)
	}
	if err := networkCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		panic(err)
	}
	if err := networkCmd.MarkPersistentFlagRequired("driver-chainid"); err != nil {
		panic(err)
	}

	infoCmd.PersistentFlags().BoolVarP(&infoFlags.Validators, "validators", "v", false, "Print compact validators info")
	infoCmd.PersistentFlags().BoolVarP(&infoFlags.ValidatorsFull, "validators-full", "V", false, "Print full validators info")
	infoCmd.PersistentFlags().BoolVarP(&infoFlags.Addresses, "addresses", "a", false, "Print addresses")
	infoCmd.PersistentFlags().BoolVarP(&infoFlags.Settlement, "settlement", "s", false, "Print settlement info")

	genesisCmd.PersistentFlags().BoolVar(&genesisFlags.Commit, "commit", false, "Commit genesis flag")
	genesisCmd.PersistentFlags().Var(&genesisFlags.Secrets, "secret-keys", "Secret key for genesis commit  in format 'chainId:key,chainId:key' (e.g. '1:0xabc,137:0xdef')")
	genesisCmd.PersistentFlags().BoolVarP(&genesisFlags.Json, "json", "j", false, "Print as json")
	genesisCmd.PersistentFlags().StringVarP(&genesisFlags.Output, "output", "o", "", "Output file path")
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		pterm.Warning.Println("Received termination signal, shutting down...")
		cancel()
	}()

	return cnCtx
}
