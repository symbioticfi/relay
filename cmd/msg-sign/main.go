package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/pkg/log"
)

func main() {
	slog.Info("Running msg sign command", "args", os.Args)

	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Msg sign completed successfully")
}

func run() error {
	rootCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&cfg.driverAddress, "driver-address", "", "Driver contract address")
	rootCmd.PersistentFlags().Uint64Var(&cfg.driverChainID, "driver-chain-id", 111, "Driver chain id")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")
	rootCmd.PersistentFlags().StringArrayVar(&cfg.signAddresses, "sign-address", []string{"http://localhost:8081/api/v1", "http://localhost:8082/api/v1", "http://localhost:8083/api/v1"}, "Addresses of signer servers'")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		return errors.Errorf("failed to mark driver-address as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL        string
	driverAddress string
	driverChainID uint64

	logLevel      string
	logMode       string
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

		driverAddress := entity.CrossChainAddress{ChainId: 111, Address: common.HexToAddress(cfg.driverAddress)}
		ethClient, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs:      []string{cfg.rpcURL},
			DriverAddress:  driverAddress,
			RequestTimeout: time.Second * 10,
		})
		if err != nil {
			return errors.Errorf("failed to create symbiotic client: %w", err)
		}
		slog.DebugContext(ctx, "Created symbiotic client")

		networkConfig, epoch, err := getLastCommittedHeaderEpoch(ctx, ethClient)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		message := strconv.FormatFloat(rand.Float64(), 'f', 10, 64) //nolint:gosec // This is just a random message for testing purposes.

		slog.DebugContext(ctx, "Trying to send sign requests", "message", message)
		reqHash, err := sendSignRequests(ctx, cfg, message, entity.ValsetHeaderKeyTag, epoch)
		if err != nil {
			return errors.Errorf("failed to send sign request: %w", err)
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				resp, err := sendGetAggregationProofRequest(ctx, cfg, reqHash)
				if err != nil {
					slog.DebugContext(ctx, "Failed to get aggregation proof", "error", err)
					continue
				}

				verifyQuorumSig(ctx, networkConfig, resp, message, ethClient, epoch)

				return nil
			case <-ctx.Done():
				slog.InfoContext(ctx, "Context canceled, stopping sign requests")
				return ctx.Err()
			}
		}
	},
}

func verifyQuorumSig(ctx context.Context, networkConfig entity.NetworkConfig, proof entity.AggregationProof, message string, eth *evm.Client, epoch uint64) {
	slog.InfoContext(ctx, "Received message with proof",
		"messageHash", hex.EncodeToString(proof.MessageHash),
		"ourMessage", hex.EncodeToString([]byte(message)),
		"ourMessageHash", hex.EncodeToString(crypto.Keccak256([]byte(message))),
	)

	ourHash := crypto.Keccak256([]byte(message))

	quorumBytes := proof.Proof[len(proof.Proof)-32:]
	quorumInt := new(big.Int).SetBytes(quorumBytes)

	for _, replica := range networkConfig.Replicas {
		verifyResult, err := eth.VerifyQuorumSig(ctx, replica, epoch, ourHash, entity.ValsetHeaderKeyTag, quorumInt, proof.Proof)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to verify quorum signature", "replica", replica.Address.Hex(), "error", err)
			continue
		}

		slog.InfoContext(ctx, "Quorum signature verification result", "result", verifyResult)
	}
}

type signMessageRequest struct {
	Message       []byte `json:"message"`
	KeyTag        uint8  `json:"keyTag"`
	RequiredEpoch uint64 `json:"requiredEpoch"`
}

func sendSignRequests(ctx context.Context, cfg config, message string, keyTag entity.KeyTag, epoch uint64) (string, error) {
	req := signMessageRequest{
		Message:       []byte(message),
		KeyTag:        uint8(keyTag),
		RequiredEpoch: epoch,
	}

	body, err := json.Marshal(&req)
	if err != nil {
		return "", errors.Errorf("failed to marshal sign message request: %w", err)
	}

	var requestHash string

	for _, signAddress := range cfg.signAddresses {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, signAddress+"/signMessage", bytes.NewReader(body))
		if err != nil {
			return "", errors.Errorf("failed to create new request: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")
		err = func() error {
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				return errors.Errorf("failed to send request to %s: %w", signAddress, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("unexpected status code from %s: %s", signAddress, resp.Status)
			}

			type response struct {
				RequestHash string `json:"requestHash"`
			}

			respData := response{}
			err = json.NewDecoder(resp.Body).Decode(&respData)
			if err != nil {
				return errors.Errorf("failed to decode response from %s: %w", signAddress, err)
			}
			requestHash = respData.RequestHash

			slog.InfoContext(ctx, "Sent sign request", "message", message, "address", signAddress, "status", resp.Status)
			return nil
		}()
		if err != nil {
			return "", errors.Errorf("failed to send request to %s: %w", signAddress, err)
		}
	}

	return requestHash, nil
}

func sendGetAggregationProofRequest(ctx context.Context, c config, hash string) (entity.AggregationProof, error) {
	url := c.signAddresses[0] + "/getAggregationProof?requestHash=" + hash
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return entity.AggregationProof{}, errors.Errorf("failed to create new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	var aggProof entity.AggregationProof
	err = func() error {
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return errors.Errorf("failed to send request to %s: %w", url, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("unexpected status code from %s: %s", url, resp.Status)
		}

		type response struct {
			VerificationType uint32 `json:"verification_type"`
			MessageHash      []byte `json:"message_hash"`
			Proof            []byte `json:"proof"`
		}

		respData := response{}
		err = json.NewDecoder(resp.Body).Decode(&respData)
		if err != nil {
			return errors.Errorf("failed to decode response from %s: %w", url, err)
		}

		aggProof = entity.AggregationProof{
			VerificationType: entity.VerificationType(respData.VerificationType),
			MessageHash:      respData.MessageHash,
			Proof:            respData.Proof,
		}

		return nil
	}()
	if err != nil {
		return entity.AggregationProof{}, errors.Errorf("failed to send request to %s: %w", url, err)
	}

	return aggProof, nil
}

func getLastCommittedHeaderEpoch(ctx context.Context, ethClient *evm.Client) (entity.NetworkConfig, uint64, error) {
	currentEpoch, err := ethClient.GetCurrentEpoch(ctx)
	if err != nil {
		return entity.NetworkConfig{}, 0, errors.Errorf("failed to get current epoch: %w", err)
	}
	epochStart, err := ethClient.GetEpochStart(ctx, currentEpoch)
	if err != nil {
		return entity.NetworkConfig{}, 0, errors.Errorf("failed to get epoch start: %w", err)
	}
	networkConfig, err := ethClient.GetConfig(ctx, epochStart)
	if err != nil {
		return entity.NetworkConfig{}, 0, errors.Errorf("failed to get network config: %w", err)
	}

	maxEpoch := uint64(0)

	for _, addr := range networkConfig.Replicas {
		epoch, err := ethClient.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return entity.NetworkConfig{}, 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch >= maxEpoch {
			maxEpoch = epoch
		}
	}

	return networkConfig, maxEpoch, nil
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("Received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
