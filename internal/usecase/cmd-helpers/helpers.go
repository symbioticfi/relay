package cmdhelpers

import (
	"fmt"
	"log/slog"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"syscall"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type SecretKeyMapFlag struct {
	Secrets map[uint64]string
}

func (s *SecretKeyMapFlag) String() string {
	parts := make([]string, 0)
	for chainID, key := range s.Secrets {
		parts = append(parts, fmt.Sprintf("%d:%s", chainID, key))
	}
	sort.Strings(parts) // Optional: consistent output order
	return strings.Join(parts, ",")
}

func (s *SecretKeyMapFlag) Set(val string) error {
	if val == "" {
		s.Secrets = make(map[uint64]string)
		return nil
	}

	result := make(map[uint64]string)
	pairs := strings.Split(val, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return errors.Errorf("invalid format (expected chainId:key): %s", pair)
		}

		chainID, err := strconv.ParseUint(kv[0], 10, 32)
		if err != nil {
			return errors.Errorf("invalid chain ID: %s", kv[0])
		}

		key := kv[1]
		if key == "" {
			return errors.Errorf("empty key for chain ID: %d", chainID)
		}

		result[chainID] = key
	}

	s.Secrets = result
	return nil
}

func (s *SecretKeyMapFlag) Type() string {
	return "secretKeyMap"
}

func GetPassword() (string, error) {
	slog.Info("Enter password: ")
	passwordBytes, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}

func PrintTreeValidator(leveledList pterm.LeveledList, validator symbiotic.Validator, totalVotingPower *big.Int) pterm.LeveledList {
	leveledList = append(leveledList, pterm.LeveledListItem{
		Level: 0,
		Text:  fmt.Sprintf("Validator: %s", validator.Operator.String()),
	})

	status := pterm.FgRed.Sprint("inactive")
	if validator.IsActive {
		status = pterm.FgGreen.Sprint("active")
	}
	leveledList = append(leveledList, pterm.LeveledListItem{
		Level: 1,
		Text:  fmt.Sprintf("Status: %s", status),
	})

	leveledList = append(leveledList, pterm.LeveledListItem{
		Level: 1,
		Text: fmt.Sprintf("Voting Power: %d (%0.3f%%)",
			validator.VotingPower.Int,
			GetPct(validator.VotingPower.Int, totalVotingPower),
		),
	})

	leveledList = append(leveledList, pterm.LeveledListItem{
		Level: 1,
		Text:  fmt.Sprintf("Vaults (%d):", len(validator.Vaults)),
	})

	for _, vault := range validator.Vaults {
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 2,
			Text:  fmt.Sprintf("Vault: %s", vault.Vault.String()),
		})
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 3,
			Text:  fmt.Sprintf("ChainID: %d", vault.ChainID),
		})
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 3,
			Text: fmt.Sprintf("Voting Power: %d (%0.3f%%)",
				vault.VotingPower,
				GetPct(vault.VotingPower.Int, validator.VotingPower.Int),
			),
		})
	}

	leveledList = append(leveledList, pterm.LeveledListItem{
		Level: 1,
		Text:  fmt.Sprintf("Keys (%d):", len(validator.Keys)),
	})

	for _, key := range validator.Keys {
		typeText, _ := key.Tag.Type().String()
		pubkeyText, _ := key.Payload.MarshalText()
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 2,
			Text:  fmt.Sprintf("Key: %d", uint8(key.Tag)),
		})
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 3,
			Text:  fmt.Sprintf("Type: %s", typeText),
		})
		leveledList = append(leveledList, pterm.LeveledListItem{
			Level: 3,
			Text:  fmt.Sprintf("PubKey: %s", pubkeyText),
		})
	}

	return leveledList
}

func GetPct(value *big.Int, total *big.Int) float64 {
	pct := new(big.Float).SetInt(value)
	pct = pct.Mul(pct, big.NewFloat(100))
	pct = pct.Quo(pct, new(big.Float).SetInt(total))
	fl, _ := pct.Float64()
	return fl
}
