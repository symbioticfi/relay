//go:build !test
// +build !test

package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"middleware-offchain/internal/client/eth"
	"middleware-offchain/valset"
	"os"

	"github.com/spf13/cobra"
)

var (
	rpcURL        string
	masterAddress string
	outputFile    string
	logLevel      string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&masterAddress, "master-address", "", "Master contract address")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		log.Fatal(err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("master-address"); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "valset",
	Short: "Generate validator set header",
	Run: func(cmd *cobra.Command, args []string) {
		// Set the log level
		var level slog.Level
		switch logLevel {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

		slog.Info("Generating valset header...")

		client, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:  rpcURL,
			MasterAddress: masterAddress,
			PrivateKey:    nil,
		})
		if err != nil {
			log.Fatalf("Failed to create eth client: %v", err)
		}

		deriver, err := valset.NewValsetDeriver(client)
		if err != nil {
			log.Fatalf("Failed to create valset deriver: %v", err)
		}

		generator, err := valset.NewValsetGenerator(deriver, client)
		if err != nil {
			log.Fatalf("Failed to create valset generator: %v", err)
		}

		header, err := generator.GenerateValidatorSetHeader(context.Background())
		if err != nil {
			log.Fatalf("Failed to generate validator set header: %v", err)
		}

		// Convert byte arrays to hex strings before JSON marshaling
		type jsonHeader struct {
			Version              uint8 `json:"version"`
			ActiveAggregatedKeys []struct {
				Tag     uint8  `json:"tag"`
				Payload string `json:"payload"` // hex string
			} `json:"activeAggregatedKeys"`
			ValidatorsSszMRoot     string   `json:"validatorsSszMRoot"` // hex string
			ExtraData              string   `json:"extraData"`
			TotalActiveVotingPower *big.Int `json:"totalActiveVotingPower"`
		}

		jsonHeaderData := jsonHeader{
			Version: header.Version,
			ActiveAggregatedKeys: make([]struct {
				Tag     uint8  `json:"tag"`
				Payload string `json:"payload"`
			}, len(header.ActiveAggregatedKeys)),
			ValidatorsSszMRoot:     fmt.Sprintf("0x%064x", header.ValidatorsSszMRoot),
			ExtraData:              FormatPayload(header.ExtraData),
			TotalActiveVotingPower: header.TotalActiveVotingPower,
		}

		for i, key := range header.ActiveAggregatedKeys {
			jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
			jsonHeaderData.ActiveAggregatedKeys[i].Payload = FormatPayload(key.Payload)
		}

		slog.Info("Valset header generated!")

		jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal header to JSON: %v", err)
		}

		if outputFile != "" {
			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				log.Fatalf("Failed to write output file: %v", err)
			}
		} else {
			fmt.Println(string(jsonData))
		}
	},
}

func FormatPayload(payload []byte) string {
	lengthHex := fmt.Sprintf("%064x", len(payload)) // 64 hex digits (32 bytes) for length
	payloadHex := hex.EncodeToString(payload)       // raw bytes → hex

	return "0x" + lengthHex + payloadHex
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
