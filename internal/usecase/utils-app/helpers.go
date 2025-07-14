package utils_app

import (
	"context"
	"fmt"
	"log/slog"
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	symbioticCrypto "middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	entity2 "middleware-offchain/internal/entity"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

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

func GetEvmClient(
	ctx context.Context,
	secretKey string,
	driver entity2.CMDCrossChainAddress,
	chainsId []uint64,
	chainsUrl []string,
) (*evm.Client, error) {
	var err error
	var privateKey []byte

	if secretKey != "" {
		privateKey = common.Hex2Bytes(secretKey)
	}

	driverCrossChainAddress := entity.CrossChainAddress{
		ChainId: driver.ChainID,
		Address: common.HexToAddress(driver.Address),
	}

	kp, err := keyprovider.NewSimpleKeystoreProvider()
	if err != nil {
		return nil, err
	}

	chains := make([]entity.ChainURL, len(chainsUrl))
	for i := range chains {
		chains[i] = entity.ChainURL{
			ChainID: chainsId[i], RPCURL: chainsUrl[i],
		}
		if len(privateKey) > 0 {
			pk, err := symbioticCrypto.NewPrivateKey(entity.KeyTypeEcdsaSecp256k1, privateKey)
			if err != nil {
				return nil, err
			}
			err = kp.AddKeyByNamespaceTypeId(keyprovider.EVM_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, int(chainsId[i]), pk)
			if err != nil {
				return nil, err
			}
		}
	}

	client, err := evm.NewEVMClient(ctx, evm.Config{
		Chains:         chains,
		DriverAddress:  driverCrossChainAddress,
		RequestTimeout: time.Second * 5,
		KeyProvider:    kp,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create symbiotic client: %w", err)
	}

	return client, nil
}
