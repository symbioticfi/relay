package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/spf13/cobra"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/log"
)

func main() {
	slog.Info("Running msg sign command", "args", os.Args)

	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Msg sign completed successfully")
}

func run() error {
	rootCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&cfg.masterAddress, "master-address", "", "Master contract address")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")
	rootCmd.PersistentFlags().StringVar(&cfg.listenAddress, "p2p-listen", "", "P2P listen address, for example '/ip4/127.0.0.1/tcp/8000'")
	rootCmd.PersistentFlags().StringArrayVar(&cfg.signAddresses, "sign-address", []string{"http://localhost:8081/api/v1/signMessage", "http://localhost:8082/api/v1/signMessage", "http://localhost:8083/api/v1/signMessage"}, "Addresses of signer servers'")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("master-address"); err != nil {
		return errors.Errorf("failed to mark master-address as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL        string
	masterAddress string

	logLevel      string
	logMode       string
	listenAddress string
	signAddresses []string
}

var cfg config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "msg-sign",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Init(cfg.logLevel, cfg.logMode)

		ctx := signalContext(context.Background())

		var opts []libp2p.Option
		if cfg.listenAddress != "" {
			opts = append(opts, libp2p.ListenAddrStrings(cfg.listenAddress))
		}
		h, err := libp2p.New(opts...)
		if err != nil {
			return errors.Errorf("failed to create libp2p host: %w", err)
		}

		p2pService, err := p2p.NewService(ctx, h)
		if err != nil {
			return errors.Errorf("failed to create p2p service: %w", err)
		}
		slog.InfoContext(ctx, "created p2p service", "listenAddr", cfg.listenAddress)
		defer p2pService.Close()

		discoveryService, err := p2p.NewDiscoveryService(ctx, p2pService, h)
		if err != nil {
			return errors.Errorf("failed to create discovery service: %w", err)
		}
		defer discoveryService.Close()
		slog.InfoContext(ctx, "created discovery service", "listenAddr", cfg.listenAddress)
		if err := discoveryService.Start(); err != nil {
			return errors.Errorf("failed to start discovery service: %w", err)
		}
		slog.InfoContext(ctx, "started discovery service", "listenAddr", cfg.listenAddress)

		ethClient, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
			RequestTimeout: time.Second * 10,
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}
		slog.DebugContext(ctx, "created eth client")

		message := strconv.FormatFloat(rand.Float64(), 'f', 10, 64) //nolint:gosec // This is just a random message for testing purposes.

		closableCtx, cancel := context.WithCancel(ctx)

		p2pService.SetSignaturesAggregatedMessageHandler(func(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
			if msg.Message.HashType == entity.HashTypeValsetHeader {
				return nil
			}

			slog.InfoContext(ctx, "received message with proof",
				"messageHash", hex.EncodeToString(msg.Message.Message),
				"ourMessage", hex.EncodeToString([]byte(message)),
				"ourMessageHash", hex.EncodeToString(crypto.Keccak256([]byte(message))),
			)

			quorumThresholdPercent := new(big.Int).Mul(big.NewInt(66), big.NewInt(1e16))
			verifyResult, err := ethClient.VerifyQuorumSig(ctx, msg.Message.Epoch, msg.Message.Message, 15, quorumThresholdPercent, msg.Message.Proof)
			if err != nil {
				return errors.Errorf("failed to verify quorum signature: %w", err)
			}

			slog.InfoContext(ctx, "quorum signature verification result", "result", verifyResult)

			cancel()

			return nil
		})
		slog.InfoContext(ctx, "sign message p2p listener created, waiting for messages")

		slog.DebugContext(ctx, "trying to send sign requests", "message", message)
		if err := sendSignRequests(ctx, cfg, message); err != nil {
			return errors.Errorf("failed to send sign request: %w", err)
		}

		<-closableCtx.Done()

		return nil
	},
}

func sendSignRequests(ctx context.Context, cfg config, message string) error {
	body := fmt.Sprintf(`{"data": "%s"}`, base64.StdEncoding.EncodeToString([]byte(message)))

	for _, signAddress := range cfg.signAddresses {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, signAddress, strings.NewReader(body))
		if err != nil {
			return errors.Errorf("failed to create new request: %w", err)
		}
		err = func() error {
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				return errors.Errorf("failed to send request to %s: %w", signAddress, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("unexpected status code from %s: %s", signAddress, resp.Status)
			}

			slog.InfoContext(ctx, "sent sign request", "message", message, "address", signAddress, "status", resp.Status)
			return nil
		}()
		if err != nil {
			return errors.Errorf("failed to send request to %s: %w", signAddress, err)
		}
	}

	return nil
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
