package utils_app

import (
	"context"
	"log/slog"
	"middleware-offchain/core/entity"
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

func LogValidator(ctx context.Context, validator entity.Validator, compact bool) error {
	status := "active"
	if !validator.IsActive {
		status = "inactive"
	}
	slog.InfoContext(ctx, "- "+validator.Operator.String())
	slog.InfoContext(ctx, "    - ", "status", status)
	slog.InfoContext(ctx, "    - ", "voting_power", validator.VotingPower)

	if compact {
		return nil
	}

	slog.InfoContext(ctx, "    -  keys")
	for _, key := range validator.Keys {
		bytes, err := key.Tag.MarshalText()
		if err != nil {
			return err
		}

		kt := string(bytes)

		bytes, err = key.Payload.MarshalText()
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "        -", "key", string(bytes), "tag", kt)
	}

	slog.InfoContext(ctx, "    -  vaults")
	for _, vault := range validator.Vaults {
		slog.InfoContext(ctx, "        -", "addr", vault.Vault, "chain_id", vault.ChainID, "voting_power", vault.VotingPower)
	}

	return nil
}
