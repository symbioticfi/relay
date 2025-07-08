package utils_app

import (
	"fmt"
	"log/slog"
	"middleware-offchain/core/entity"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func GetPassword() (string, error) {
	slog.Info("Enter password: ")
	passwordBytes, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	return string(passwordBytes), nil
}

func MarshalTextValidator(validator entity.Validator, compact bool) (string, error) {
	var result strings.Builder

	status := "active"
	if !validator.IsActive {
		status = "inactive"
	}

	result.WriteString(fmt.Sprintf("\nValidator: %s\n", validator.Operator.String()))
	result.WriteString(fmt.Sprintf("   Status: %s\n", status))
	result.WriteString(fmt.Sprintf("   Voting Power: %v\n", validator.VotingPower))

	if compact {
		return result.String(), nil
	}

	result.WriteString(fmt.Sprintf("\nKeys (%d):\n", len(validator.Keys)))
	result.WriteString("   # | Key | Tag\n")
	for i, key := range validator.Keys {
		tagBytes, err := key.Tag.MarshalText()
		if err != nil {
			return "", err
		}

		payloadBytes, err := key.Payload.MarshalText()
		if err != nil {
			return "", err
		}

		result.WriteString(fmt.Sprintf("   %d | %s | %s\n", i+1, string(payloadBytes), string(tagBytes)))
	}

	result.WriteString(fmt.Sprintf("\nVaults (%d):\n", len(validator.Vaults)))
	result.WriteString("   # | Address | Chain ID | Voting Power\n")
	for i, vault := range validator.Vaults {
		result.WriteString(fmt.Sprintf("   %d | %s | %d | %v\n", i+1, vault.Vault, vault.ChainID, vault.VotingPower))
	}

	return result.String(), nil
}
