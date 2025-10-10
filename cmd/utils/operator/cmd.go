package operator

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var operatorRegistries = map[uint64]symbiotic.CrossChainAddress{
	111: {
		ChainId: 111,
		Address: common.HexToAddress("0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"),
	},
}

func NewOperatorCmd() *cobra.Command {
	operatorCmd.AddCommand(infoCmd)
	operatorCmd.AddCommand(registerCmd)
	operatorCmd.AddCommand(registerKeyCmd)

	initFlags()

	return operatorCmd
}

var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Operator tool",
}

type GlobalFlags struct {
	Chains        []string
	DriverAddress string
	DriverChainId uint64
}

type InfoFlags struct {
	Epoch    uint64
	Full     bool
	Path     string
	Password string
	KeyTag   uint8
}

type RegisterFlags struct {
	Secrets cmdhelpers.SecretKeyMapFlag
}

type RegisterKeyFlags struct {
	Secrets  cmdhelpers.SecretKeyMapFlag
	Path     string
	Password string
	KeyTag   uint8
}

var globalFlags GlobalFlags
var infoFlags InfoFlags
var registerFlags RegisterFlags
var registerKeyFlags RegisterKeyFlags

func initFlags() {
	operatorCmd.PersistentFlags().StringSliceVarP(&globalFlags.Chains, "chains", "c", nil, "Chains rpc url, comma separated")
	operatorCmd.PersistentFlags().StringVar(&globalFlags.DriverAddress, "driver-address", "", "Driver contract address")
	operatorCmd.PersistentFlags().Uint64Var(&globalFlags.DriverChainId, "driver-chainid", 0, "Driver contract chain id")
	if err := operatorCmd.MarkPersistentFlagRequired("chains"); err != nil {
		panic(err)
	}
	if err := operatorCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		panic(err)
	}
	if err := operatorCmd.MarkPersistentFlagRequired("driver-chainid"); err != nil {
		panic(err)
	}

	infoCmd.PersistentFlags().Uint64VarP(&infoFlags.Epoch, "epoch", "e", 0, "Network epoch to fetch info")
	infoCmd.PersistentFlags().BoolVarP(&infoFlags.Full, "full", "f", false, "Print full validator info")
	infoCmd.PersistentFlags().StringVarP(&infoFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	infoCmd.PersistentFlags().StringVar(&infoFlags.Password, "password", "", "Keystore password")
	infoCmd.PersistentFlags().Uint8Var(&infoFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag")
	if err := infoCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}

	registerCmd.PersistentFlags().Var(&registerFlags.Secrets, "secret-keys", "Secret key for register in format 'chainId:key,chainId:key' (e.g. '1:0xabc,137:0xdef')")

	registerKeyCmd.PersistentFlags().Var(&registerFlags.Secrets, "secret-keys", "Secret key for key register in format 'chainId:key' (e.g. '1:0xabc')")
	registerKeyCmd.PersistentFlags().StringVarP(&infoFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	registerKeyCmd.PersistentFlags().StringVar(&infoFlags.Password, "password", "", "Keystore password")
	registerKeyCmd.PersistentFlags().Uint8Var(&infoFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag")
	if err := registerKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}
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
