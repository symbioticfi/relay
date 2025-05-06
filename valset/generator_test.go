package valset

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
)

func TestGenerator(t *testing.T) {
	// Define the large number
	privateKeyInt := new(big.Int)
	privateKeyInt.SetString("87191036493798670866484781455694320176667203290824056510541300741498740913410", 10)

	client, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:  "http://127.0.0.1:8545",
		MasterAddress: "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:    lo.ToPtr(privateKeyInt.Bytes()),
	})
	require.NoError(t, err)

	deriver, err := NewValsetDeriver(client)
	require.NoError(t, err)

	generator, err := NewValsetGenerator(deriver, client)
	require.NoError(t, err)

	header, err := generator.GenerateValidatorSetHeader(context.Background())
	require.NoError(t, err)

	// Convert byte arrays to hex strings before JSON marshaling
	type jsonHeader struct {
		ActiveAggregatedKeys []struct {
			Tag     uint8  `json:"tag"`
			Payload string `json:"payload"` // hex string
		} `json:"activeAggregatedKeys"`
		ValidatorsSszMRoot string `json:"validatorsSszMRoot"` // hex string
	}

	jsonHeaderData := jsonHeader{
		ActiveAggregatedKeys: make([]struct {
			Tag     uint8  `json:"tag"`
			Payload string `json:"payload"`
		}, len(header.ActiveAggregatedKeys)),
		ValidatorsSszMRoot: fmt.Sprintf("0x%064x%064x", len(header.ValidatorsSszMRoot), header.ValidatorsSszMRoot),
	}

	for i, key := range header.ActiveAggregatedKeys {
		jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
		jsonHeaderData.ActiveAggregatedKeys[i].Payload = fmt.Sprintf("0x%064x%064x", len(key.Payload), key.Payload)
	}

	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal header to JSON: %v", err)
	}
	slog.Debug("Generated validator set header", "json", string(jsonData))

	//fmt.Println(hex.EncodeToString(encoded))
}
