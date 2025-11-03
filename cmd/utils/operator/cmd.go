package operator

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewOperatorCmd() *cobra.Command {
	operatorCmd.AddCommand(infoCmd)
	operatorCmd.AddCommand(registerKeyCmd)
	operatorCmd.AddCommand(registerSignatureCmd)
	operatorCmd.AddCommand(unregisterSignatureCmd)

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

type RegisterKeyFlags struct {
	Secrets  cmdhelpers.SecretKeyMapFlag
	Path     string
	Password string
	KeyTag   uint8
}

type RegisterSignatureFlags struct {
	Secrets cmdhelpers.SecretKeyMapFlag
}

type UnregisterSignatureFlags struct {
	Secrets cmdhelpers.SecretKeyMapFlag
}

var globalFlags GlobalFlags
var infoFlags InfoFlags
var registerKeyFlags RegisterKeyFlags
var registerSignatureFlags RegisterSignatureFlags
var unregisterSignatureFlags UnregisterSignatureFlags

func initFlags() {
	operatorCmd.PersistentFlags().StringSliceVarP(&globalFlags.Chains, "chains", "c", nil, "Chains rpc url, comma separated")
	operatorCmd.PersistentFlags().StringVar(&globalFlags.DriverAddress, "driver.address", "", "Driver contract address")
	operatorCmd.PersistentFlags().Uint64Var(&globalFlags.DriverChainId, "driver.chainid", 0, "Driver contract chain id")
	if err := operatorCmd.MarkPersistentFlagRequired("chains"); err != nil {
		panic(err)
	}
	if err := operatorCmd.MarkPersistentFlagRequired("driver.address"); err != nil {
		panic(err)
	}
	if err := operatorCmd.MarkPersistentFlagRequired("driver.chainid"); err != nil {
		panic(err)
	}

	infoCmd.PersistentFlags().Uint64VarP(&infoFlags.Epoch, "epoch", "e", 0, "Network epoch to fetch info")
	infoCmd.PersistentFlags().StringVarP(&infoFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	infoCmd.PersistentFlags().StringVar(&infoFlags.Password, "password", "", "Keystore password")
	infoCmd.PersistentFlags().Uint8Var(&infoFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag")
	if err := infoCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}

	registerKeyCmd.PersistentFlags().Var(&registerKeyFlags.Secrets, "secret-keys", "Secret key for key register in format 'chainId:key' (e.g. '1:0xabc')")
	registerKeyCmd.PersistentFlags().StringVarP(&registerKeyFlags.Path, "path", "p", "./keystore.jks", "Path to keystore")
	registerKeyCmd.PersistentFlags().StringVar(&registerKeyFlags.Password, "password", "", "Keystore password")
	registerKeyCmd.PersistentFlags().Uint8Var(&registerKeyFlags.KeyTag, "key-tag", uint8(symbiotic.KeyTypeInvalid), "key tag")
	if err := registerKeyCmd.MarkPersistentFlagRequired("key-tag"); err != nil {
		panic(err)
	}

	registerSignatureCmd.PersistentFlags().Var(&registerSignatureFlags.Secrets, "secret-keys", "Secret key for signing in format 'chainId:key' (e.g. '1:0xabc')")

	unregisterSignatureCmd.PersistentFlags().Var(&unregisterSignatureFlags.Secrets, "secret-keys", "Secret key for signing in format 'chainId:key' (e.g. '1:0xabc')")
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
